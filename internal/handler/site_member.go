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

type SiteMemberHandler struct {
	memberRepo *repository.SiteMemberRepository
	userRepo   *repository.AdminUserRepository
	log        *slog.Logger
}

func NewSiteMemberHandler(memberRepo *repository.SiteMemberRepository, userRepo *repository.AdminUserRepository, log *slog.Logger) *SiteMemberHandler {
	return &SiteMemberHandler{memberRepo: memberRepo, userRepo: userRepo, log: log}
}

func (h *SiteMemberHandler) List(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	members, err := h.memberRepo.ListBySite(r.Context(), siteID)
	if err != nil {
		h.log.Error("listing site members", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to list members")
		return
	}

	response.JSON(w, http.StatusOK, members)
}

func (h *SiteMemberHandler) Add(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	var req model.AddSiteMemberRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" {
		response.Error(w, http.StatusBadRequest, "user_id is required")
		return
	}

	userOID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	if _, err := h.userRepo.FindByID(r.Context(), userOID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("finding user for membership", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to add member")
		return
	}

	member := model.SiteMember{
		SiteID: siteID,
		UserID: userOID,
		Role:   model.SiteRoleViewer,
	}

	created, err := h.memberRepo.Create(r.Context(), member)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			response.Error(w, http.StatusConflict, "user is already a member of this site")
			return
		}
		h.log.Error("adding site member", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to add member")
		return
	}

	response.JSON(w, http.StatusCreated, created)
}

func (h *SiteMemberHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	memberID, err := primitive.ObjectIDFromHex(r.PathValue("memberId"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid member ID")
		return
	}

	var req model.UpdateSiteMemberRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Role == "" {
		response.Error(w, http.StatusBadRequest, "role is required")
		return
	}

	if _, ok := model.ValidSiteRoles[req.Role]; !ok {
		response.Error(w, http.StatusBadRequest, "role must be one of: owner, editor, viewer")
		return
	}

	existing, err := h.memberRepo.FindByID(r.Context(), memberID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "member not found")
			return
		}
		h.log.Error("finding member for update", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to update member")
		return
	}

	if existing.SiteID != siteID {
		response.Error(w, http.StatusNotFound, "member not found")
		return
	}

	if existing.Role == model.SiteRoleOwner && req.Role != model.SiteRoleOwner {
		count, err := h.memberRepo.CountOwners(r.Context(), siteID)
		if err != nil {
			h.log.Error("counting owners", "error", err, "request_id", middleware.GetRequestID(r.Context()))
			response.Error(w, http.StatusInternalServerError, "failed to update member")
			return
		}
		if count <= 1 {
			response.Error(w, http.StatusBadRequest, "cannot remove the last owner")
			return
		}
	}

	updated, err := h.memberRepo.UpdateRole(r.Context(), memberID, req.Role)
	if err != nil {
		h.log.Error("updating member role", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to update member")
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

func (h *SiteMemberHandler) Remove(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	memberID, err := primitive.ObjectIDFromHex(r.PathValue("memberId"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid member ID")
		return
	}

	existing, err := h.memberRepo.FindByID(r.Context(), memberID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "member not found")
			return
		}
		h.log.Error("finding member for delete", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to remove member")
		return
	}

	if existing.SiteID != siteID {
		response.Error(w, http.StatusNotFound, "member not found")
		return
	}

	if existing.Role == model.SiteRoleOwner {
		count, err := h.memberRepo.CountOwners(r.Context(), siteID)
		if err != nil {
			h.log.Error("counting owners", "error", err, "request_id", middleware.GetRequestID(r.Context()))
			response.Error(w, http.StatusInternalServerError, "failed to remove member")
			return
		}
		if count <= 1 {
			response.Error(w, http.StatusBadRequest, "cannot remove the last owner")
			return
		}
	}

	if err := h.memberRepo.Delete(r.Context(), memberID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "member not found")
			return
		}
		h.log.Error("removing site member", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to remove member")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
