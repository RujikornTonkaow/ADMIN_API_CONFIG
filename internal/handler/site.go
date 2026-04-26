package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"portfolio-admin-api/internal/database"
	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/model"
	"portfolio-admin-api/internal/repository"
	"portfolio-admin-api/pkg/response"
)

type SiteHandler struct {
	siteRepo   *repository.SiteRepository
	memberRepo *repository.SiteMemberRepository
	db         *mongo.Database
	log        *slog.Logger
}

func NewSiteHandler(siteRepo *repository.SiteRepository, memberRepo *repository.SiteMemberRepository, db *mongo.Database, log *slog.Logger) *SiteHandler {
	return &SiteHandler{siteRepo: siteRepo, memberRepo: memberRepo, db: db, log: log}
}

func (h *SiteHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userRole := middleware.GetRole(ctx)

	userIDHex := middleware.GetUserID(ctx)
	userOID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid user identity")
		return
	}

	if userRole == model.RoleSuperAdmin {
		sites, err := h.siteRepo.List(ctx)
		if err != nil {
			h.log.Error("listing all sites", "error", err, "request_id", middleware.GetRequestID(ctx))
			response.Error(w, http.StatusInternalServerError, "failed to list sites")
			return
		}
		h.ensurePortfolioContent(ctx, sites)
		response.JSON(w, http.StatusOK, sites)
		return
	}

	members, err := h.memberRepo.ListByUser(ctx, userOID)
	if err != nil {
		h.log.Error("listing user memberships", "error", err, "request_id", middleware.GetRequestID(ctx))
		response.Error(w, http.StatusInternalServerError, "failed to list sites")
		return
	}

	siteIDs := make([]primitive.ObjectID, len(members))
	for i, m := range members {
		siteIDs[i] = m.SiteID
	}

	if len(siteIDs) == 0 {
		response.JSON(w, http.StatusOK, []model.Site{})
		return
	}

	sites, err := h.siteRepo.ListByIDs(ctx, siteIDs)
	if err != nil {
		h.log.Error("listing sites by ids", "error", err, "request_id", middleware.GetRequestID(ctx))
		response.Error(w, http.StatusInternalServerError, "failed to list sites")
		return
	}

	h.ensurePortfolioContent(ctx, sites)
	response.JSON(w, http.StatusOK, sites)
}

func (h *SiteHandler) ensurePortfolioContent(ctx context.Context, sites []model.Site) {
	for _, site := range sites {
		if site.Type != model.SiteTypePortfolio {
			continue
		}
		if err := database.EnsurePortfolioContent(ctx, h.db, site.ID); err != nil {
			h.log.Error("ensuring portfolio content", "error", err, "site_id", site.ID.Hex(), "request_id", middleware.GetRequestID(ctx))
		}
	}
}

func (h *SiteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSiteRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Slug == "" || req.Type == "" {
		response.Error(w, http.StatusBadRequest, "name, slug, and type are required")
		return
	}

	if !model.ValidSiteTypes[req.Type] {
		response.Error(w, http.StatusBadRequest, "type must be one of: portfolio, shop, finance")
		return
	}

	if req.Domains == nil {
		req.Domains = []string{}
	}

	userIDHex := middleware.GetUserID(r.Context())
	userOID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid user identity")
		return
	}

	site := model.Site{
		Name:    req.Name,
		Slug:    req.Slug,
		Type:    req.Type,
		Domains: req.Domains,
	}

	created, err := h.siteRepo.Create(r.Context(), site)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			response.Error(w, http.StatusConflict, "site slug already exists")
			return
		}
		h.log.Error("creating site", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to create site")
		return
	}

	member := model.SiteMember{
		SiteID: created.ID,
		UserID: userOID,
		Role:   model.SiteRoleViewer,
	}
	if _, err := h.memberRepo.Create(r.Context(), member); err != nil {
		h.log.Error("creating owner membership", "error", err, "request_id", middleware.GetRequestID(r.Context()))
	}

	if req.Type == model.SiteTypePortfolio {
		if err := database.SeedPortfolioContent(r.Context(), h.db, created.ID); err != nil {
			h.log.Error("seeding portfolio content", "error", err, "request_id", middleware.GetRequestID(r.Context()))
			response.Error(w, http.StatusInternalServerError, "failed to seed portfolio content")
			return
		}
	}

	response.JSON(w, http.StatusCreated, created)
}

func (h *SiteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	site, err := h.siteRepo.FindByID(r.Context(), siteID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "site not found")
			return
		}
		h.log.Error("getting site", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to get site")
		return
	}

	response.JSON(w, http.StatusOK, site)
}

func (h *SiteHandler) Update(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	var req model.UpdateSiteRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Slug == "" {
		response.Error(w, http.StatusBadRequest, "name and slug are required")
		return
	}

	if req.Domains == nil {
		req.Domains = []string{}
	}

	updated, err := h.siteRepo.Update(r.Context(), siteID, req)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "site not found")
			return
		}
		if mongo.IsDuplicateKeyError(err) {
			response.Error(w, http.StatusConflict, "site slug already exists")
			return
		}
		h.log.Error("updating site", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to update site")
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

func (h *SiteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	siteID := middleware.GetSiteID(r.Context())

	if err := h.siteRepo.Delete(r.Context(), siteID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "site not found")
			return
		}
		h.log.Error("deleting site", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to delete site")
		return
	}

	if err := h.memberRepo.DeleteBySite(r.Context(), siteID); err != nil {
		h.log.Error("cleaning up site members", "error", err, "request_id", middleware.GetRequestID(r.Context()))
	}

	w.WriteHeader(http.StatusNoContent)
}
