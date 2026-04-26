package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"

	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/model"
	"portfolio-admin-api/internal/repository"
	"portfolio-admin-api/pkg/response"
)

type SiteSettingsHandler struct {
	repo *repository.SiteSettingsRepository
	log  *slog.Logger
}

func NewSiteSettingsHandler(repo *repository.SiteSettingsRepository, log *slog.Logger) *SiteSettingsHandler {
	return &SiteSettingsHandler{repo: repo, log: log}
}

func (h *SiteSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	settings, err := h.repo.Get(r.Context(), siteID)
	if err != nil {
		h.log.Error("getting site settings",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to get site settings")
		return
	}
	response.JSON(w, http.StatusOK, settings)
}

func (h *SiteSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	var req model.SiteSettings
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SiteTitle == "" {
		existing, err := h.repo.Get(r.Context(), siteID)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				response.Error(w, http.StatusBadRequest, "site_title is required")
				return
			}
			h.log.Error("getting existing site settings",
				"error", err,
				"request_id", middleware.GetRequestID(r.Context()),
			)
			response.Error(w, http.StatusInternalServerError, "failed to get site settings")
			return
		}
		req = mergeSiteSettings(existing, req)
	}

	result, err := h.repo.Upsert(r.Context(), siteID, req)
	if err != nil {
		h.log.Error("updating site settings",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to update site settings")
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func mergeSiteSettings(existing, update model.SiteSettings) model.SiteSettings {
	if update.SiteTitle == "" {
		update.SiteTitle = existing.SiteTitle
	}
	if update.PageTitle == "" {
		update.PageTitle = existing.PageTitle
	}
	if update.MetaDescription == "" {
		update.MetaDescription = existing.MetaDescription
	}
	if update.FooterTagline == "" {
		update.FooterTagline = existing.FooterTagline
	}
	if update.DefaultTheme == "" {
		update.DefaultTheme = existing.DefaultTheme
	}
	if update.ProfileImage == "" {
		update.ProfileImage = existing.ProfileImage
	}
	return update
}
