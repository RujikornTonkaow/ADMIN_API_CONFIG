package handler

import (
	"log/slog"
	"net/http"

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
	settings, err := h.repo.Get(r.Context())
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
	var req model.SiteSettings
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SiteTitle == "" {
		response.Error(w, http.StatusBadRequest, "site_title is required")
		return
	}

	result, err := h.repo.Upsert(r.Context(), req)
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
