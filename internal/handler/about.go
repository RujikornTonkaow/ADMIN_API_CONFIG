package handler

import (
	"log/slog"
	"net/http"

	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/model"
	"portfolio-admin-api/internal/repository"
	"portfolio-admin-api/pkg/response"
)

type AboutHandler struct {
	repo *repository.AboutRepository
	log  *slog.Logger
}

func NewAboutHandler(repo *repository.AboutRepository, log *slog.Logger) *AboutHandler {
	return &AboutHandler{repo: repo, log: log}
}

func (h *AboutHandler) Get(w http.ResponseWriter, r *http.Request) {
	about, err := h.repo.Get(r.Context())
	if err != nil {
		h.log.Error("getting about",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to get about data")
		return
	}
	response.JSON(w, http.StatusOK, about)
}

func (h *AboutHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.About
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		response.Error(w, http.StatusBadRequest, "title is required")
		return
	}

	result, err := h.repo.Upsert(r.Context(), req)
	if err != nil {
		h.log.Error("updating about",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to update about data")
		return
	}
	response.JSON(w, http.StatusOK, result)
}
