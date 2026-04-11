package router

import (
	"log/slog"
	"net/http"

	"portfolio-admin-api/internal/handler"
	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/model"
	"portfolio-admin-api/internal/repository"
)

type Config struct {
	JWTSecret      string
	UploadDir      string
	MaxUploadMB    int64
	AllowedOrigins string
	Log            *slog.Logger

	SiteSettingsRepo *repository.SiteSettingsRepository
	HeroRepo         *repository.HeroRepository
	AboutRepo        *repository.AboutRepository
	SkillRepo        *repository.SkillRepository
	ProjectRepo      *repository.ProjectRepository
	ExperienceRepo   *repository.ExperienceRepository
	SocialLinkRepo   *repository.SocialLinkRepository
	ContactRepo      *repository.ContactRepository
	AdminUserRepo    *repository.AdminUserRepository
}

func New(cfg *Config) http.Handler {
	mux := http.NewServeMux()

	authMw := middleware.Auth(cfg.JWTSecret)
	requireAdmin := middleware.RequireRole(model.RoleAdmin, authMw)
	requireUser := middleware.RequireRole(model.RoleUserAccount, authMw)
	requireVisitor := middleware.RequireRole(model.RoleVisitor, authMw)

	authH := handler.NewAuthHandler(cfg.AdminUserRepo, cfg.JWTSecret, cfg.Log)
	publicH := handler.NewPublicHandler(
		cfg.SiteSettingsRepo, cfg.HeroRepo, cfg.AboutRepo,
		cfg.SkillRepo, cfg.ProjectRepo, cfg.ExperienceRepo,
		cfg.SocialLinkRepo, cfg.ContactRepo, cfg.Log,
	)
	siteSettingsH := handler.NewSiteSettingsHandler(cfg.SiteSettingsRepo, cfg.Log)
	heroH := handler.NewHeroHandler(cfg.HeroRepo, cfg.Log)
	aboutH := handler.NewAboutHandler(cfg.AboutRepo, cfg.Log)
	skillH := handler.NewSkillHandler(cfg.SkillRepo, cfg.Log)
	projectH := handler.NewProjectHandler(cfg.ProjectRepo, cfg.Log)
	experienceH := handler.NewExperienceHandler(cfg.ExperienceRepo, cfg.Log)
	socialLinkH := handler.NewSocialLinkHandler(cfg.SocialLinkRepo, cfg.Log)
	contactH := handler.NewContactHandler(cfg.ContactRepo, cfg.Log)
	uploadH := handler.NewUploadHandler(cfg.UploadDir, cfg.MaxUploadMB, cfg.Log)
	userH := handler.NewUserHandler(cfg.AdminUserRepo, cfg.Log)

	// --- Public API (no auth) ---
	mux.HandleFunc("GET /api/v1/portfolio", publicH.GetPortfolio)
	mux.HandleFunc("POST /api/v1/contact", publicH.SubmitContact)

	// --- Auth ---
	mux.HandleFunc("POST /api/v1/admin/auth/login", authH.Login)
	mux.HandleFunc("GET /api/v1/admin/auth/me", requireVisitor(authH.GetMe))

	// --- Admin: User Management (admin only) ---
	mux.HandleFunc("GET /api/v1/admin/users", requireAdmin(userH.List))
	mux.HandleFunc("POST /api/v1/admin/users", requireAdmin(userH.Create))
	mux.HandleFunc("GET /api/v1/admin/users/{id}", requireAdmin(userH.GetByID))
	mux.HandleFunc("PUT /api/v1/admin/users/{id}", requireAdmin(userH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/users/{id}", requireAdmin(userH.Delete))
	mux.HandleFunc("PUT /api/v1/admin/users/{id}/password", requireAdmin(userH.ChangePassword))

	// --- Admin: Site Settings (user_account+) ---
	mux.HandleFunc("GET /api/v1/admin/site-settings", requireUser(siteSettingsH.Get))
	mux.HandleFunc("PUT /api/v1/admin/site-settings", requireUser(siteSettingsH.Update))

	// --- Admin: Hero (user_account+) ---
	mux.HandleFunc("GET /api/v1/admin/hero", requireUser(heroH.Get))
	mux.HandleFunc("PUT /api/v1/admin/hero", requireUser(heroH.Update))

	// --- Admin: About (user_account+) ---
	mux.HandleFunc("GET /api/v1/admin/about", requireUser(aboutH.Get))
	mux.HandleFunc("PUT /api/v1/admin/about", requireUser(aboutH.Update))

	// --- Admin: Skills (user_account+) ---
	mux.HandleFunc("GET /api/v1/admin/skills", requireUser(skillH.List))
	mux.HandleFunc("POST /api/v1/admin/skills", requireUser(skillH.Create))
	mux.HandleFunc("PUT /api/v1/admin/skills/{id}", requireUser(skillH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/skills/{id}", requireUser(skillH.Delete))

	// --- Admin: Projects (user_account+) ---
	mux.HandleFunc("GET /api/v1/admin/projects", requireUser(projectH.List))
	mux.HandleFunc("POST /api/v1/admin/projects", requireUser(projectH.Create))
	mux.HandleFunc("PUT /api/v1/admin/projects/reorder", requireUser(projectH.Reorder))
	mux.HandleFunc("PUT /api/v1/admin/projects/{id}", requireUser(projectH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/projects/{id}", requireUser(projectH.Delete))

	// --- Admin: Experience (user_account+) ---
	mux.HandleFunc("GET /api/v1/admin/experiences", requireUser(experienceH.List))
	mux.HandleFunc("POST /api/v1/admin/experiences", requireUser(experienceH.Create))
	mux.HandleFunc("PUT /api/v1/admin/experiences/reorder", requireUser(experienceH.Reorder))
	mux.HandleFunc("PUT /api/v1/admin/experiences/{id}", requireUser(experienceH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/experiences/{id}", requireUser(experienceH.Delete))

	// --- Admin: Social Links (user_account+) ---
	mux.HandleFunc("GET /api/v1/admin/social-links", requireUser(socialLinkH.List))
	mux.HandleFunc("POST /api/v1/admin/social-links", requireUser(socialLinkH.Create))
	mux.HandleFunc("PUT /api/v1/admin/social-links/reorder", requireUser(socialLinkH.Reorder))
	mux.HandleFunc("PUT /api/v1/admin/social-links/{id}", requireUser(socialLinkH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/social-links/{id}", requireUser(socialLinkH.Delete))

	// --- Admin: Contact Messages (read: visitor+, delete: user_account+) ---
	mux.HandleFunc("GET /api/v1/admin/contacts", requireVisitor(contactH.List))
	mux.HandleFunc("GET /api/v1/admin/contacts/{id}", requireVisitor(contactH.GetByID))
	mux.HandleFunc("DELETE /api/v1/admin/contacts/{id}", requireUser(contactH.Delete))

	// --- Admin: File Upload (user_account+) ---
	mux.HandleFunc("POST /api/v1/admin/upload", requireUser(uploadH.Upload))

	// --- Static files: uploaded images ---
	fs := http.FileServer(http.Dir(cfg.UploadDir))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	// --- Global middleware stack ---
	return middleware.Chain(mux,
		middleware.Recovery(cfg.Log),
		middleware.CORS(cfg.AllowedOrigins),
		middleware.Logging(cfg.Log),
		middleware.RequestID,
	)
}
