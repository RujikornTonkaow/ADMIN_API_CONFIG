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

type ExperienceHandler struct {
	repo *repository.ExperienceRepository
	log  *slog.Logger
}

func NewExperienceHandler(repo *repository.ExperienceRepository, log *slog.Logger) *ExperienceHandler {
	return &ExperienceHandler{repo: repo, log: log}
}

func (h *ExperienceHandler) List(w http.ResponseWriter, r *http.Request) {
	experiences, err := h.repo.List(r.Context())
	if err != nil {
		h.log.Error("listing experiences",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to list experiences")
		return
	}
	response.JSON(w, http.StatusOK, experiences)
}

func (h *ExperienceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.Experience
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Role == "" || req.Company == "" || req.Period == "" {
		response.Error(w, http.StatusBadRequest, "role, company, and period are required")
		return
	}

	if req.Highlights == nil {
		req.Highlights = []string{}
	}

	created, err := h.repo.Create(r.Context(), req)
	if err != nil {
		h.log.Error("creating experience",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to create experience")
		return
	}
	response.JSON(w, http.StatusCreated, created)
}

func (h *ExperienceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid experience ID")
		return
	}

	var req model.Experience
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Role == "" || req.Company == "" || req.Period == "" {
		response.Error(w, http.StatusBadRequest, "role, company, and period are required")
		return
	}

	updated, err := h.repo.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "experience not found")
			return
		}
		h.log.Error("updating experience",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to update experience")
		return
	}
	response.JSON(w, http.StatusOK, updated)
}

func (h *ExperienceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid experience ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "experience not found")
			return
		}
		h.log.Error("deleting experience",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to delete experience")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ExperienceHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	var req model.ReorderRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.IDs) == 0 {
		response.Error(w, http.StatusBadRequest, "ids array is required")
		return
	}

	objectIDs := make([]primitive.ObjectID, len(req.IDs))
	for i, idStr := range req.IDs {
		oid, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid ID in array: "+idStr)
			return
		}
		objectIDs[i] = oid
	}

	if err := h.repo.Reorder(r.Context(), objectIDs); err != nil {
		h.log.Error("reordering experiences",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to reorder experiences")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}
