package router

import (
	"log/slog"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"

	"portfolio-admin-api/internal/handler"
	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/model"
	"portfolio-admin-api/internal/repository"
)

type Config struct {
	JWTSecret   string
	UploadDir   string
	MaxUploadMB int64
	DomainCache *middleware.DomainCache
	Log         *slog.Logger
	DB          *mongo.Database

	SiteSettingsRepo *repository.SiteSettingsRepository
	HeroRepo         *repository.HeroRepository
	AboutRepo        *repository.AboutRepository
	SkillRepo        *repository.SkillRepository
	ProjectRepo      *repository.ProjectRepository
	ExperienceRepo   *repository.ExperienceRepository
	SocialLinkRepo   *repository.SocialLinkRepository
	ContactRepo      *repository.ContactRepository
	AdminUserRepo    *repository.AdminUserRepository
	SiteRepo         *repository.SiteRepository
	SiteMemberRepo   *repository.SiteMemberRepository
}

func New(cfg *Config) http.Handler {
	mux := http.NewServeMux()

	authMw := middleware.Auth(cfg.JWTSecret)
	requireSuperAdmin := middleware.RequireRole(model.RoleSuperAdmin, authMw)
	requireAdmin := middleware.RequireRole(model.RoleAdmin, authMw)
	requireViewer := middleware.RequireRole(model.RoleViewer, authMw)

	siteAdmin := middleware.RequireSiteMember(cfg.SiteMemberRepo, model.RoleAdmin, authMw)
	siteEditor := middleware.RequireSiteMember(cfg.SiteMemberRepo, model.RoleEditor, authMw)
	siteViewer := middleware.RequireSiteMember(cfg.SiteMemberRepo, model.RoleViewer, authMw)

	authH := handler.NewAuthHandler(cfg.AdminUserRepo, cfg.JWTSecret, cfg.Log)
	publicH := handler.NewPublicHandler(
		cfg.SiteSettingsRepo, cfg.HeroRepo, cfg.AboutRepo,
		cfg.SkillRepo, cfg.ProjectRepo, cfg.ExperienceRepo,
		cfg.SocialLinkRepo, cfg.ContactRepo, cfg.SiteRepo, cfg.Log,
	)
	siteH := handler.NewSiteHandler(cfg.SiteRepo, cfg.SiteMemberRepo, cfg.DB, cfg.Log)
	siteMemberH := handler.NewSiteMemberHandler(cfg.SiteMemberRepo, cfg.AdminUserRepo, cfg.Log)
	siteSettingsH := handler.NewSiteSettingsHandler(cfg.SiteSettingsRepo, cfg.Log)
	heroH := handler.NewHeroHandler(cfg.HeroRepo, cfg.Log)
	aboutH := handler.NewAboutHandler(cfg.AboutRepo, cfg.Log)
	skillH := handler.NewSkillHandler(cfg.SkillRepo, cfg.Log)
	projectH := handler.NewProjectHandler(cfg.ProjectRepo, cfg.Log)
	experienceH := handler.NewExperienceHandler(cfg.ExperienceRepo, cfg.Log)
	socialLinkH := handler.NewSocialLinkHandler(cfg.SocialLinkRepo, cfg.Log)
	contactH := handler.NewContactHandler(cfg.ContactRepo, cfg.Log)
	uploadH := handler.NewUploadHandler(cfg.UploadDir, cfg.MaxUploadMB, cfg.Log)
	userH := handler.NewUserHandler(cfg.AdminUserRepo, cfg.SiteMemberRepo, cfg.Log)

	// --- Public API (no auth) ---
	mux.HandleFunc("GET /api/v1/public/sites/by-domain", publicH.ResolveDomain)
	mux.HandleFunc("GET /api/v1/public/sites/{siteId}/portfolio", publicH.GetPortfolio)
	mux.HandleFunc("POST /api/v1/public/sites/{siteId}/portfolio/contacts", publicH.SubmitContact)

	// --- Auth ---
	mux.HandleFunc("POST /api/v1/admin/auth/login", authH.Login)
	mux.HandleFunc("GET /api/v1/admin/auth/me", requireViewer(authH.GetMe))

	// --- Admin: User Management (admin+, scoped in handler) ---
	mux.HandleFunc("GET /api/v1/admin/users", requireAdmin(userH.List))
	mux.HandleFunc("POST /api/v1/admin/users", requireAdmin(userH.Create))
	mux.HandleFunc("GET /api/v1/admin/users/{id}", requireAdmin(userH.GetByID))
	mux.HandleFunc("PUT /api/v1/admin/users/{id}", requireAdmin(userH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/users/{id}", requireAdmin(userH.Delete))
	mux.HandleFunc("PUT /api/v1/admin/users/{id}/password", requireAdmin(userH.ChangePassword))
	mux.HandleFunc("GET /api/v1/admin/users/{id}/memberships", requireAdmin(userH.ListMemberships))
	mux.HandleFunc("PUT /api/v1/admin/users/{id}/memberships", requireAdmin(userH.UpdateMemberships))

	// --- Site Management ---
	mux.HandleFunc("GET /api/v1/admin/sites", requireViewer(siteH.List))
	mux.HandleFunc("POST /api/v1/admin/sites", requireSuperAdmin(siteH.Create))
	mux.HandleFunc("GET /api/v1/admin/sites/{siteId}", siteViewer(siteH.GetByID))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}", siteAdmin(siteH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/sites/{siteId}", siteAdmin(siteH.Delete))

	// --- Site Members (admin+) ---
	mux.HandleFunc("GET /api/v1/admin/sites/{siteId}/members", siteAdmin(siteMemberH.List))
	mux.HandleFunc("POST /api/v1/admin/sites/{siteId}/members", siteAdmin(siteMemberH.Add))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}/members/{memberId}", siteAdmin(siteMemberH.UpdateRole))
	mux.HandleFunc("DELETE /api/v1/admin/sites/{siteId}/members/{memberId}", siteAdmin(siteMemberH.Remove))

	// --- Portfolio: Site Settings (site member, editor+) ---
	mux.HandleFunc("GET /api/v1/admin/sites/{siteId}/portfolio/site-settings", siteEditor(siteSettingsH.Get))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}/portfolio/site-settings", siteEditor(siteSettingsH.Update))

	// --- Portfolio: Hero (site member, editor+) ---
	mux.HandleFunc("GET /api/v1/admin/sites/{siteId}/portfolio/hero", siteEditor(heroH.Get))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}/portfolio/hero", siteEditor(heroH.Update))

	// --- Portfolio: About (site member, editor+) ---
	mux.HandleFunc("GET /api/v1/admin/sites/{siteId}/portfolio/about", siteEditor(aboutH.Get))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}/portfolio/about", siteEditor(aboutH.Update))

	// --- Portfolio: Skills (site member, editor+) ---
	mux.HandleFunc("GET /api/v1/admin/sites/{siteId}/portfolio/skills", siteEditor(skillH.List))
	mux.HandleFunc("POST /api/v1/admin/sites/{siteId}/portfolio/skills", siteEditor(skillH.Create))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}/portfolio/skills/{id}", siteEditor(skillH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/sites/{siteId}/portfolio/skills/{id}", siteEditor(skillH.Delete))

	// --- Portfolio: Projects (site member, editor+) ---
	mux.HandleFunc("GET /api/v1/admin/sites/{siteId}/portfolio/projects", siteEditor(projectH.List))
	mux.HandleFunc("POST /api/v1/admin/sites/{siteId}/portfolio/projects", siteEditor(projectH.Create))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}/portfolio/projects/reorder", siteEditor(projectH.Reorder))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}/portfolio/projects/{id}", siteEditor(projectH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/sites/{siteId}/portfolio/projects/{id}", siteEditor(projectH.Delete))

	// --- Portfolio: Experiences (site member, editor+) ---
	mux.HandleFunc("GET /api/v1/admin/sites/{siteId}/portfolio/experiences", siteEditor(experienceH.List))
	mux.HandleFunc("POST /api/v1/admin/sites/{siteId}/portfolio/experiences", siteEditor(experienceH.Create))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}/portfolio/experiences/reorder", siteEditor(experienceH.Reorder))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}/portfolio/experiences/{id}", siteEditor(experienceH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/sites/{siteId}/portfolio/experiences/{id}", siteEditor(experienceH.Delete))

	// --- Portfolio: Social Links (site member, editor+) ---
	mux.HandleFunc("GET /api/v1/admin/sites/{siteId}/portfolio/social-links", siteEditor(socialLinkH.List))
	mux.HandleFunc("POST /api/v1/admin/sites/{siteId}/portfolio/social-links", siteEditor(socialLinkH.Create))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}/portfolio/social-links/reorder", siteEditor(socialLinkH.Reorder))
	mux.HandleFunc("PUT /api/v1/admin/sites/{siteId}/portfolio/social-links/{id}", siteEditor(socialLinkH.Update))
	mux.HandleFunc("DELETE /api/v1/admin/sites/{siteId}/portfolio/social-links/{id}", siteEditor(socialLinkH.Delete))

	// --- Portfolio: Contact Messages (site member, viewer+/editor+) ---
	mux.HandleFunc("GET /api/v1/admin/sites/{siteId}/portfolio/contacts", siteViewer(contactH.List))
	mux.HandleFunc("GET /api/v1/admin/sites/{siteId}/portfolio/contacts/{id}", siteViewer(contactH.GetByID))
	mux.HandleFunc("DELETE /api/v1/admin/sites/{siteId}/portfolio/contacts/{id}", siteEditor(contactH.Delete))

	// --- Portfolio: File Upload (site member, editor+) ---
	mux.HandleFunc("POST /api/v1/admin/sites/{siteId}/portfolio/upload", siteEditor(uploadH.Upload))

	// --- Static files: uploaded images ---
	fs := http.FileServer(http.Dir(cfg.UploadDir))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	// --- Global middleware stack ---
	return middleware.Chain(mux,
		middleware.Recovery(cfg.Log),
		middleware.CORS(cfg.DomainCache),
		middleware.Logging(cfg.Log),
		middleware.RequestID,
	)
}
