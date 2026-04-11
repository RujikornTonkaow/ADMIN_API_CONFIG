package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/model"
	"portfolio-admin-api/internal/repository"
	"portfolio-admin-api/pkg/response"
)

func isValidEmail(email string) bool {
	at := strings.Index(email, "@")
	if at < 1 {
		return false
	}
	dot := strings.LastIndex(email[at:], ".")
	return dot > 1 && dot < len(email[at:])-1
}

type PublicHandler struct {
	siteSettings *repository.SiteSettingsRepository
	hero         *repository.HeroRepository
	about        *repository.AboutRepository
	skills       *repository.SkillRepository
	projects     *repository.ProjectRepository
	experiences  *repository.ExperienceRepository
	socialLinks  *repository.SocialLinkRepository
	contacts     *repository.ContactRepository
	log          *slog.Logger
}

func NewPublicHandler(
	siteSettings *repository.SiteSettingsRepository,
	hero *repository.HeroRepository,
	about *repository.AboutRepository,
	skills *repository.SkillRepository,
	projects *repository.ProjectRepository,
	experiences *repository.ExperienceRepository,
	socialLinks *repository.SocialLinkRepository,
	contacts *repository.ContactRepository,
	log *slog.Logger,
) *PublicHandler {
	return &PublicHandler{
		siteSettings: siteSettings,
		hero:         hero,
		about:        about,
		skills:       skills,
		projects:     projects,
		experiences:  experiences,
		socialLinks:  socialLinks,
		contacts:     contacts,
		log:          log,
	}
}

func (h *PublicHandler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetRequestID(ctx)

	settings, err := h.siteSettings.Get(ctx)
	if err != nil {
		h.log.Error("fetching site settings for portfolio", "error", err, "request_id", reqID)
	}

	heroData, err := h.hero.Get(ctx)
	if err != nil {
		h.log.Error("fetching hero for portfolio", "error", err, "request_id", reqID)
	}

	aboutData, err := h.about.Get(ctx)
	if err != nil {
		h.log.Error("fetching about for portfolio", "error", err, "request_id", reqID)
	}

	skills, err := h.skills.List(ctx)
	if err != nil {
		h.log.Error("fetching skills for portfolio", "error", err, "request_id", reqID)
		skills = []model.Skill{}
	}

	projects, err := h.projects.List(ctx)
	if err != nil {
		h.log.Error("fetching projects for portfolio", "error", err, "request_id", reqID)
		projects = []model.Project{}
	}

	experiences, err := h.experiences.List(ctx)
	if err != nil {
		h.log.Error("fetching experiences for portfolio", "error", err, "request_id", reqID)
		experiences = []model.Experience{}
	}

	socialLinks, err := h.socialLinks.List(ctx)
	if err != nil {
		h.log.Error("fetching social links for portfolio", "error", err, "request_id", reqID)
		socialLinks = []model.SocialLink{}
	}

	navItems := []model.NavItem{
		{Label: "Home", Href: "#hero"},
		{Label: "About", Href: "#about"},
		{Label: "Skills", Href: "#skills"},
		{Label: "Projects", Href: "#projects"},
		{Label: "Experience", Href: "#experience"},
		{Label: "Contact", Href: "#contact"},
	}

	portfolio := model.PortfolioData{
		SiteSettings: settings,
		Hero:         heroData,
		About:        aboutData,
		Skills:       skills,
		Projects:     projects,
		Experiences:  experiences,
		SocialLinks:  socialLinks,
		NavItems:     navItems,
	}

	response.JSON(w, http.StatusOK, portfolio)
}

func (h *PublicHandler) SubmitContact(w http.ResponseWriter, r *http.Request) {
	var req model.ContactRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Email == "" || req.Subject == "" || req.Message == "" {
		response.Error(w, http.StatusBadRequest, "name, email, subject, and message are required")
		return
	}

	if !isValidEmail(req.Email) {
		response.Error(w, http.StatusBadRequest, "invalid email format")
		return
	}

	msg := model.ContactMessage{
		Name:    req.Name,
		Email:   req.Email,
		Subject: req.Subject,
		Message: req.Message,
	}

	if _, err := h.contacts.Create(r.Context(), msg); err != nil {
		h.log.Error("creating contact message",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to submit message")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "Contact message sent successfully"})
}
