package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/model"
	"portfolio-admin-api/internal/repository"
	"portfolio-admin-api/pkg/response"
)

type UserHandler struct {
	repo       *repository.AdminUserRepository
	memberRepo *repository.SiteMemberRepository
	log        *slog.Logger
}

func NewUserHandler(repo *repository.AdminUserRepository, memberRepo *repository.SiteMemberRepository, log *slog.Logger) *UserHandler {
	return &UserHandler{repo: repo, memberRepo: memberRepo, log: log}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.List(r.Context())
	if err != nil {
		h.log.Error("listing users", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	if middleware.GetRole(r.Context()) != model.RoleSuperAdmin {
		users, err = h.filterManageableUsers(r, users)
		if err != nil {
			h.log.Error("filtering manageable users", "error", err, "request_id", middleware.GetRequestID(r.Context()))
			response.Error(w, http.StatusInternalServerError, "failed to list users")
			return
		}
	}

	response.JSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getManageableUser(w, r)
	if !ok {
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" || req.Role == "" {
		response.Error(w, http.StatusBadRequest, "username, password, and role are required")
		return
	}
	if _, ok := model.ValidRoles[req.Role]; !ok {
		response.Error(w, http.StatusBadRequest, "role must be super_admin, admin, editor, or viewer")
		return
	}
	if !h.canAssignRole(middleware.GetRole(r.Context()), req.Role) {
		response.Error(w, http.StatusForbidden, "cannot assign this role")
		return
	}
	if len(req.Password) < 8 {
		response.Error(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("hashing password", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	created, err := h.repo.Create(r.Context(), model.AdminUser{
		Username: req.Username,
		Password: string(hashed),
		Role:     req.Role,
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			response.Error(w, http.StatusConflict, "username already exists")
			return
		}
		h.log.Error("creating user", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	response.JSON(w, http.StatusCreated, created)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req model.UpdateUserRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" || req.Role == "" {
		response.Error(w, http.StatusBadRequest, "username and role are required")
		return
	}
	if _, ok := model.ValidRoles[req.Role]; !ok {
		response.Error(w, http.StatusBadRequest, "role must be super_admin, admin, editor, or viewer")
		return
	}

	existing, ok := h.getManageableUserByID(w, r, id)
	if !ok {
		return
	}
	currentRole := middleware.GetRole(r.Context())
	if id.Hex() == middleware.GetUserID(r.Context()) && existing.Role != req.Role {
		response.Error(w, http.StatusBadRequest, "cannot change your own role")
		return
	}
	if !h.canAssignRole(currentRole, req.Role) {
		response.Error(w, http.StatusForbidden, "cannot assign this role")
		return
	}
	if existing.Role == model.RoleSuperAdmin && req.Role != model.RoleSuperAdmin && !h.hasAnotherSuperAdmin(r, existing.ID) {
		response.Error(w, http.StatusBadRequest, "cannot remove the last super admin")
		return
	}

	updated, err := h.repo.Update(r.Context(), id, req.Username, req.Role)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			response.Error(w, http.StatusConflict, "username already exists")
			return
		}
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("updating user", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to update user")
		return
	}
	response.JSON(w, http.StatusOK, updated)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user ID")
		return
	}
	if id.Hex() == middleware.GetUserID(r.Context()) {
		response.Error(w, http.StatusBadRequest, "cannot delete your own account")
		return
	}

	target, ok := h.getManageableUserByID(w, r, id)
	if !ok {
		return
	}
	if target.Role == model.RoleSuperAdmin && !h.hasAnotherSuperAdmin(r, target.ID) {
		response.Error(w, http.StatusBadRequest, "cannot delete the last super admin")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("deleting user", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req model.ChangePasswordRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.NewPassword) < 8 {
		response.Error(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if _, ok := h.getManageableUserByID(w, r, id); !ok {
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("hashing new password", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to change password")
		return
	}
	if err := h.repo.UpdatePassword(r.Context(), id, string(hashed)); err != nil {
		h.log.Error("updating password", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to change password")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "password changed"})
}

func (h *UserHandler) ListMemberships(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getManageableUser(w, r)
	if !ok {
		return
	}

	memberships, err := h.memberRepo.ListByUser(r.Context(), user.ID)
	if err != nil {
		h.log.Error("listing user memberships", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to list memberships")
		return
	}

	response.JSON(w, http.StatusOK, membershipResponses(memberships))
}

func (h *UserHandler) UpdateMemberships(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getManageableUser(w, r)
	if !ok {
		return
	}
	if user.Role == model.RoleSuperAdmin {
		response.Error(w, http.StatusBadRequest, "super admins do not need site memberships")
		return
	}

	var req model.UpdateUserMembershipsRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	desired := make(map[primitive.ObjectID]struct{}, len(req.Memberships))
	for _, membership := range req.Memberships {
		siteID, err := primitive.ObjectIDFromHex(membership.SiteID)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid site_id")
			return
		}
		if !h.canManageSiteAccess(r, siteID) {
			response.Error(w, http.StatusForbidden, "cannot manage access for this site")
			return
		}
		desired[siteID] = struct{}{}
	}

	existing, err := h.memberRepo.ListByUser(r.Context(), user.ID)
	if err != nil {
		h.log.Error("listing existing memberships", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to update memberships")
		return
	}

	for _, current := range existing {
		if _, keep := desired[current.SiteID]; keep {
			delete(desired, current.SiteID)
			continue
		}
		if !h.canManageSiteAccess(r, current.SiteID) {
			continue
		}
		if err := h.memberRepo.DeleteBySiteAndUser(r.Context(), current.SiteID, user.ID); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			h.log.Error("deleting user membership", "error", err, "request_id", middleware.GetRequestID(r.Context()))
			response.Error(w, http.StatusInternalServerError, "failed to update memberships")
			return
		}
	}

	for siteID := range desired {
		if _, err := h.memberRepo.Create(r.Context(), model.SiteMember{SiteID: siteID, UserID: user.ID, Role: model.SiteRoleViewer}); err != nil {
			if mongo.IsDuplicateKeyError(err) {
				continue
			}
			h.log.Error("creating user membership", "error", err, "request_id", middleware.GetRequestID(r.Context()))
			response.Error(w, http.StatusInternalServerError, "failed to update memberships")
			return
		}
	}

	updated, err := h.memberRepo.ListByUser(r.Context(), user.ID)
	if err != nil {
		h.log.Error("listing updated memberships", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to update memberships")
		return
	}
	response.JSON(w, http.StatusOK, membershipResponses(updated))
}

func (h *UserHandler) getManageableUser(w http.ResponseWriter, r *http.Request) (model.AdminUser, bool) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user ID")
		return model.AdminUser{}, false
	}
	return h.getManageableUserByID(w, r, id)
}

func (h *UserHandler) getManageableUserByID(w http.ResponseWriter, r *http.Request, id primitive.ObjectID) (model.AdminUser, bool) {
	user, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "user not found")
			return model.AdminUser{}, false
		}
		h.log.Error("finding user", "error", err, "request_id", middleware.GetRequestID(r.Context()))
		response.Error(w, http.StatusInternalServerError, "failed to get user")
		return model.AdminUser{}, false
	}
	if !h.canManageUser(r, user) {
		response.Error(w, http.StatusForbidden, "cannot manage this user")
		return model.AdminUser{}, false
	}
	return user, true
}

func (h *UserHandler) canManageUser(r *http.Request, target model.AdminUser) bool {
	currentRole := middleware.GetRole(r.Context())
	if currentRole == model.RoleSuperAdmin {
		return true
	}
	if currentRole != model.RoleAdmin || target.Role == model.RoleSuperAdmin {
		return false
	}
	return h.usersShareManagedSite(r, target.ID)
}

func (h *UserHandler) canAssignRole(currentRole, targetRole string) bool {
	if currentRole == model.RoleSuperAdmin {
		return true
	}
	return currentRole == model.RoleAdmin && targetRole != model.RoleSuperAdmin
}

func (h *UserHandler) filterManageableUsers(r *http.Request, users []model.AdminUser) ([]model.AdminUser, error) {
	filtered := make([]model.AdminUser, 0, len(users))
	for _, user := range users {
		if user.ID.Hex() == middleware.GetUserID(r.Context()) || h.canManageUser(r, user) {
			filtered = append(filtered, user)
		}
	}
	return filtered, nil
}

func (h *UserHandler) usersShareManagedSite(r *http.Request, targetID primitive.ObjectID) bool {
	currentID, err := primitive.ObjectIDFromHex(middleware.GetUserID(r.Context()))
	if err != nil {
		return false
	}

	currentMemberships, err := h.memberRepo.ListByUser(r.Context(), currentID)
	if err != nil {
		return false
	}
	if len(currentMemberships) == 0 {
		return false
	}

	targetMemberships, err := h.memberRepo.ListByUser(r.Context(), targetID)
	if err != nil {
		return false
	}
	if len(targetMemberships) == 0 {
		return true
	}

	currentSites := make(map[primitive.ObjectID]struct{}, len(currentMemberships))
	for _, membership := range currentMemberships {
		currentSites[membership.SiteID] = struct{}{}
	}
	for _, membership := range targetMemberships {
		if _, ok := currentSites[membership.SiteID]; ok {
			return true
		}
	}
	return false
}

func (h *UserHandler) canManageSiteAccess(r *http.Request, siteID primitive.ObjectID) bool {
	if middleware.GetRole(r.Context()) == model.RoleSuperAdmin {
		return true
	}
	if middleware.GetRole(r.Context()) != model.RoleAdmin {
		return false
	}
	currentUserID, err := primitive.ObjectIDFromHex(middleware.GetUserID(r.Context()))
	if err != nil {
		return false
	}
	_, err = h.memberRepo.FindBySiteAndUser(r.Context(), siteID, currentUserID)
	return err == nil
}

func (h *UserHandler) hasAnotherSuperAdmin(r *http.Request, currentID primitive.ObjectID) bool {
	users, err := h.repo.List(r.Context())
	if err != nil {
		return false
	}
	for _, user := range users {
		if user.ID != currentID && user.Role == model.RoleSuperAdmin {
			return true
		}
	}
	return false
}

func membershipResponses(memberships []model.SiteMember) []model.UserMembershipResponse {
	res := make([]model.UserMembershipResponse, len(memberships))
	for i, membership := range memberships {
		res[i] = model.UserMembershipResponse{
			ID:     membership.ID.Hex(),
			SiteID: membership.SiteID.Hex(),
			UserID: membership.UserID.Hex(),
		}
	}
	return res
}
