package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/repository"
	"portfolio-admin-api/pkg/response"
)

type ContactHandler struct {
	repo *repository.ContactRepository
	log  *slog.Logger
}

func NewContactHandler(repo *repository.ContactRepository, log *slog.Logger) *ContactHandler {
	return &ContactHandler{repo: repo, log: log}
}

func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	messages, err := h.repo.List(r.Context(), siteID)
	if err != nil {
		h.log.Error("listing contact messages",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to list contact messages")
		return
	}

	unread, err := h.repo.CountUnread(r.Context(), siteID)
	if err != nil {
		h.log.Error("counting unread messages",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
	}

	response.JSONWithMeta(w, http.StatusOK, messages, map[string]int64{"unread_count": unread})
}

func (h *ContactHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid message ID")
		return
	}

	msg, err := h.repo.GetByID(r.Context(), siteID, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "message not found")
			return
		}
		h.log.Error("getting contact message",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to get message")
		return
	}

	if !msg.IsRead {
		if err := h.repo.MarkAsRead(r.Context(), siteID, id); err != nil {
			h.log.Error("marking message as read",
				"error", err,
				"request_id", middleware.GetRequestID(r.Context()),
			)
		}
		msg.IsRead = true
	}

	response.JSON(w, http.StatusOK, msg)
}

func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid message ID")
		return
	}

	if err := h.repo.Delete(r.Context(), siteID, id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "message not found")
			return
		}
		h.log.Error("deleting contact message",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to delete message")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
