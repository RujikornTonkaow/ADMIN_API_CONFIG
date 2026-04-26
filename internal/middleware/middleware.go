package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"portfolio-admin-api/internal/model"
	"portfolio-admin-api/internal/repository"
	"portfolio-admin-api/pkg/response"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	UsernameKey  contextKey = "username"
	RoleKey      contextKey = "role"
	SiteIDKey    contextKey = "site_id"
)

func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

func Logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(ww, r)

			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", GetRequestID(r.Context()),
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func Recovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic recovered",
						"error", err,
						"path", r.URL.Path,
						"request_id", GetRequestID(r.Context()),
					)
					response.Error(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type DomainCache struct {
	siteRepo      *repository.SiteRepository
	staticOrigins []string
	mu            sync.RWMutex
	dynamicHosts  map[string]bool
	log           *slog.Logger
}

func NewDomainCache(siteRepo *repository.SiteRepository, staticOrigins string, log *slog.Logger) *DomainCache {
	origins := strings.Split(staticOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	dc := &DomainCache{
		siteRepo:      siteRepo,
		staticOrigins: origins,
		dynamicHosts:  make(map[string]bool),
		log:           log,
	}
	dc.Refresh()
	return dc
}

func (dc *DomainCache) Refresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	domains, err := dc.siteRepo.ListAllDomains(ctx)
	if err != nil {
		dc.log.Error("refreshing domain cache", "error", err)
		return
	}

	hosts := make(map[string]bool, len(domains))
	for _, d := range domains {
		hosts[d] = true
	}

	dc.mu.Lock()
	dc.dynamicHosts = hosts
	dc.mu.Unlock()

	dc.log.Info("domain cache refreshed", "count", len(hosts))
}

func (dc *DomainCache) StartAutoRefresh(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dc.Refresh()
			}
		}
	}()
}

func (dc *DomainCache) IsAllowed(origin string) bool {
	for _, o := range dc.staticOrigins {
		if o == origin || o == "*" {
			return true
		}
	}

	host := strings.TrimPrefix(origin, "http://")
	host = strings.TrimPrefix(host, "https://")

	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.dynamicHosts[host]
}

func CORS(dc *DomainCache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if dc.IsAllowed(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Auth(jwtSecret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				response.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			if !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}
			tokenStr := strings.TrimPrefix(header, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				response.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.Error(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			userID, _ := claims["sub"].(string)
			username, _ := claims["usr"].(string)
			role, _ := claims["role"].(string)

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, userID)
			ctx = context.WithValue(ctx, UsernameKey, username)
			ctx = context.WithValue(ctx, RoleKey, role)

			next(w, r.WithContext(ctx))
		}
	}
}

func RequireRole(minRole string, authMw func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return authMw(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetRole(r.Context())
			if model.RoleLevel(userRole) < model.RoleLevel(minRole) {
				response.Error(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next(w, r)
		})
	}
}

func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

func GetUsername(ctx context.Context) string {
	if u, ok := ctx.Value(UsernameKey).(string); ok {
		return u
	}
	return ""
}

func GetRole(ctx context.Context) string {
	if r, ok := ctx.Value(RoleKey).(string); ok {
		return r
	}
	return ""
}

func GetSiteID(ctx context.Context) primitive.ObjectID {
	if id, ok := ctx.Value(SiteIDKey).(primitive.ObjectID); ok {
		return id
	}
	return primitive.NilObjectID
}

func RequireSiteMember(memberRepo *repository.SiteMemberRepository, minRole string, authMw func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return authMw(func(w http.ResponseWriter, r *http.Request) {
			siteIDHex := r.PathValue("siteId")
			siteID, err := primitive.ObjectIDFromHex(siteIDHex)
			if err != nil {
				response.Error(w, http.StatusBadRequest, "invalid site ID")
				return
			}

			userRole := GetRole(r.Context())
			if userRole == model.RoleSuperAdmin {
				ctx := context.WithValue(r.Context(), SiteIDKey, siteID)
				next(w, r.WithContext(ctx))
				return
			}

			if model.RoleLevel(userRole) < model.RoleLevel(minRole) {
				response.Error(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			userIDHex := GetUserID(r.Context())
			userOID, err := primitive.ObjectIDFromHex(userIDHex)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid user identity")
				return
			}

			_, err = memberRepo.FindBySiteAndUser(r.Context(), siteID, userOID)
			if err != nil {
				response.Error(w, http.StatusForbidden, "you are not a member of this site")
				return
			}

			ctx := context.WithValue(r.Context(), SiteIDKey, siteID)
			next(w, r.WithContext(ctx))
		})
	}
}
