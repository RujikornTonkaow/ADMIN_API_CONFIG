package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/model"
	"portfolio-admin-api/internal/repository"
	"portfolio-admin-api/pkg/response"
)

type SkillHandler struct {
	repo *repository.SkillRepository
	log  *slog.Logger
}

func NewSkillHandler(repo *repository.SkillRepository, log *slog.Logger) *SkillHandler {
	return &SkillHandler{repo: repo, log: log}
}

func (h *SkillHandler) List(w http.ResponseWriter, r *http.Request) {
	skills, err := h.repo.List(r.Context())
	if err != nil {
		h.log.Error("listing skills",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to list skills")
		return
	}
	response.JSON(w, http.StatusOK, skills)
}

func (h *SkillHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.Skill
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Icon == "" || req.Category == "" {
		response.Error(w, http.StatusBadRequest, "name, icon, and category are required")
		return
	}

	validCategories := map[string]bool{"frontend": true, "backend": true, "devops": true, "tools": true}
	if !validCategories[req.Category] {
		response.Error(w, http.StatusBadRequest, "category must be one of: frontend, backend, devops, tools")
		return
	}

	created, err := h.repo.Create(r.Context(), req)
	if err != nil {
		h.log.Error("creating skill",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to create skill")
		return
	}
	response.JSON(w, http.StatusCreated, created)
}

func (h *SkillHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid skill ID")
		return
	}

	var req model.Skill
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Icon == "" || req.Category == "" {
		response.Error(w, http.StatusBadRequest, "name, icon, and category are required")
		return
	}

	updated, err := h.repo.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "skill not found")
			return
		}
		h.log.Error("updating skill",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to update skill")
		return
	}
	response.JSON(w, http.StatusOK, updated)
}

func (h *SkillHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid skill ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "skill not found")
			return
		}
		h.log.Error("deleting skill",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to delete skill")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
