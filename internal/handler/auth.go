package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/model"
	"portfolio-admin-api/internal/repository"
	"portfolio-admin-api/pkg/response"
)

type AuthHandler struct {
	repo      *repository.AdminUserRepository
	jwtSecret string
	log       *slog.Logger
}

func NewAuthHandler(repo *repository.AdminUserRepository, jwtSecret string, log *slog.Logger) *AuthHandler {
	return &AuthHandler{repo: repo, jwtSecret: jwtSecret, log: log}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "username and password are required")
		return
	}

	user, err := h.repo.FindByUsername(r.Context(), req.Username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		h.log.Error("finding user for login",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Hex(),
		"usr": user.Username,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenStr, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		h.log.Error("signing JWT token",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, model.LoginResponse{Token: tokenStr})
}
