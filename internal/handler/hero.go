package handler

import (
	"log/slog"
	"net/http"

	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/model"
	"portfolio-admin-api/internal/repository"
	"portfolio-admin-api/pkg/response"
)

type HeroHandler struct {
	repo *repository.HeroRepository
	log  *slog.Logger
}

func NewHeroHandler(repo *repository.HeroRepository, log *slog.Logger) *HeroHandler {
	return &HeroHandler{repo: repo, log: log}
}

func (h *HeroHandler) Get(w http.ResponseWriter, r *http.Request) {
	hero, err := h.repo.Get(r.Context())
	if err != nil {
		h.log.Error("getting hero",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to get hero data")
		return
	}
	response.JSON(w, http.StatusOK, hero)
}

func (h *HeroHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.Hero
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FullName == "" {
		response.Error(w, http.StatusBadRequest, "full_name is required")
		return
	}

	result, err := h.repo.Upsert(r.Context(), req)
	if err != nil {
		h.log.Error("updating hero",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to update hero data")
		return
	}
	response.JSON(w, http.StatusOK, result)
}
