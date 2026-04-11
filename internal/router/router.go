package router

import (
	"log/slog"
	"net/http"

	"portfolio-admin-api/internal/handler"
	"portfolio-admin-api/internal/middleware"
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

	// --- Public API (no auth) ---
	mux.HandleFunc("GET /api/v1/portfolio", publicH.GetPortfolio)
	mux.HandleFunc("POST /api/v1/contact", publicH.SubmitContact)

	// --- Auth ---
	mux.HandleFunc("POST /api/v1/admin/auth/login", authH.Login)

	// --- Admin: Site Settings (singleton) ---
	mux.HandleFunc("GET /api/v1/admin/site-settings", authMw(siteSettingsH.Get))
	mux.HandleFunc("PUT /api/v1/admin/site-settings", authMw(siteSettingsH.Update))

	// --- Admin: Hero (singleton) ---
	mux.HandleFunc("GET /api/v1/admin/hero", authMw(heroH.Get))
	mux.HandleFunc("PUT /api/v1/admin/hero", authMw(heroH.Update))

	// --- Admin: About (singleton) ---
	mux.HandleFunc("GET /api/v1/admin/about", authMw(aboutH.Get))
	mux.HandleFunc("PUT /api/v1/admin/about", authMw(aboutH.Update))

	// --- Admin: Skills (CRUD) ---
	mux.HandleFunc("GET /api/v1/admin/skills", authMw(skillH.List))
	mux.HandleFunc("POST /api/v1/admin/skills", authMw(skillH.Create))
	mux.HandleFunc("PUT /api/v1/admin/skills/{id}", authMw(skillH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/skills/{id}", authMw(skillH.Delete))

	// --- Admin: Projects (CRUD + reorder) ---
	mux.HandleFunc("GET /api/v1/admin/projects", authMw(projectH.List))
	mux.HandleFunc("POST /api/v1/admin/projects", authMw(projectH.Create))
	mux.HandleFunc("PUT /api/v1/admin/projects/reorder", authMw(projectH.Reorder))
	mux.HandleFunc("PUT /api/v1/admin/projects/{id}", authMw(projectH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/projects/{id}", authMw(projectH.Delete))

	// --- Admin: Experience (CRUD + reorder) ---
	mux.HandleFunc("GET /api/v1/admin/experiences", authMw(experienceH.List))
	mux.HandleFunc("POST /api/v1/admin/experiences", authMw(experienceH.Create))
	mux.HandleFunc("PUT /api/v1/admin/experiences/reorder", authMw(experienceH.Reorder))
	mux.HandleFunc("PUT /api/v1/admin/experiences/{id}", authMw(experienceH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/experiences/{id}", authMw(experienceH.Delete))

	// --- Admin: Social Links (CRUD + reorder) ---
	mux.HandleFunc("GET /api/v1/admin/social-links", authMw(socialLinkH.List))
	mux.HandleFunc("POST /api/v1/admin/social-links", authMw(socialLinkH.Create))
	mux.HandleFunc("PUT /api/v1/admin/social-links/reorder", authMw(socialLinkH.Reorder))
	mux.HandleFunc("PUT /api/v1/admin/social-links/{id}", authMw(socialLinkH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/social-links/{id}", authMw(socialLinkH.Delete))

	// --- Admin: Contact Messages ---
	mux.HandleFunc("GET /api/v1/admin/contacts", authMw(contactH.List))
	mux.HandleFunc("GET /api/v1/admin/contacts/{id}", authMw(contactH.GetByID))
	mux.HandleFunc("DELETE /api/v1/admin/contacts/{id}", authMw(contactH.Delete))

	// --- Admin: File Upload ---
	mux.HandleFunc("POST /api/v1/admin/upload", authMw(uploadH.Upload))

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
