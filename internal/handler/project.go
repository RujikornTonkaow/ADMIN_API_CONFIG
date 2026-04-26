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

type ProjectHandler struct {
	repo *repository.ProjectRepository
	log  *slog.Logger
}

func NewProjectHandler(repo *repository.ProjectRepository, log *slog.Logger) *ProjectHandler {
	return &ProjectHandler{repo: repo, log: log}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	projects, err := h.repo.List(r.Context(), siteID)
	if err != nil {
		h.log.Error("listing projects",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	response.JSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	var req model.Project
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" || req.Description == "" {
		response.Error(w, http.StatusBadRequest, "title and description are required")
		return
	}

	if req.Tags == nil {
		req.Tags = []string{}
	}

	created, err := h.repo.Create(r.Context(), siteID, req)
	if err != nil {
		h.log.Error("creating project",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to create project")
		return
	}
	response.JSON(w, http.StatusCreated, created)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	var req model.Project
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" || req.Description == "" {
		response.Error(w, http.StatusBadRequest, "title and description are required")
		return
	}

	updated, err := h.repo.Update(r.Context(), siteID, id, req)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "project not found")
			return
		}
		h.log.Error("updating project",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to update project")
		return
	}
	response.JSON(w, http.StatusOK, updated)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	if err := h.repo.Delete(r.Context(), siteID, id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "project not found")
			return
		}
		h.log.Error("deleting project",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to delete project")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

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

	if err := h.repo.Reorder(r.Context(), siteID, objectIDs); err != nil {
		h.log.Error("reordering projects",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to reorder projects")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}
