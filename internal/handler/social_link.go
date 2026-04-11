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

type SocialLinkHandler struct {
	repo *repository.SocialLinkRepository
	log  *slog.Logger
}

func NewSocialLinkHandler(repo *repository.SocialLinkRepository, log *slog.Logger) *SocialLinkHandler {
	return &SocialLinkHandler{repo: repo, log: log}
}

func (h *SocialLinkHandler) List(w http.ResponseWriter, r *http.Request) {
	links, err := h.repo.List(r.Context())
	if err != nil {
		h.log.Error("listing social links",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to list social links")
		return
	}
	response.JSON(w, http.StatusOK, links)
}

func (h *SocialLinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.SocialLink
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.URL == "" || req.Icon == "" {
		response.Error(w, http.StatusBadRequest, "name, url, and icon are required")
		return
	}

	created, err := h.repo.Create(r.Context(), req)
	if err != nil {
		h.log.Error("creating social link",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to create social link")
		return
	}
	response.JSON(w, http.StatusCreated, created)
}

func (h *SocialLinkHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid social link ID")
		return
	}

	var req model.SocialLink
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.URL == "" || req.Icon == "" {
		response.Error(w, http.StatusBadRequest, "name, url, and icon are required")
		return
	}

	updated, err := h.repo.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "social link not found")
			return
		}
		h.log.Error("updating social link",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to update social link")
		return
	}
	response.JSON(w, http.StatusOK, updated)
}

func (h *SocialLinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid social link ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "social link not found")
			return
		}
		h.log.Error("deleting social link",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to delete social link")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SocialLinkHandler) Reorder(w http.ResponseWriter, r *http.Request) {
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
		h.log.Error("reordering social links",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to reorder social links")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}
