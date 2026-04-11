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
	repo *repository.AdminUserRepository
	log  *slog.Logger
}

func NewUserHandler(repo *repository.AdminUserRepository, log *slog.Logger) *UserHandler {
	return &UserHandler{repo: repo, log: log}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.List(r.Context())
	if err != nil {
		h.log.Error("listing users",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	response.JSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("getting user",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to get user")
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
		response.Error(w, http.StatusBadRequest, "role must be admin, user_account, or visitor")
		return
	}

	if len(req.Password) < 8 {
		response.Error(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("hashing password",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	user := model.AdminUser{
		Username: req.Username,
		Password: string(hashed),
		Role:     req.Role,
	}

	created, err := h.repo.Create(r.Context(), user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			response.Error(w, http.StatusConflict, "username already exists")
			return
		}
		h.log.Error("creating user",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
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
		response.Error(w, http.StatusBadRequest, "role must be admin, user_account, or visitor")
		return
	}

	currentUserID := middleware.GetUserID(r.Context())
	if id.Hex() == currentUserID && req.Role != model.RoleAdmin {
		response.Error(w, http.StatusBadRequest, "cannot downgrade your own role")
		return
	}

	existing, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("finding user for update",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	if existing.Role == model.RoleAdmin && req.Role != model.RoleAdmin {
		count, err := h.repo.CountByRole(r.Context(), model.RoleAdmin)
		if err != nil {
			h.log.Error("counting admins",
				"error", err,
				"request_id", middleware.GetRequestID(r.Context()),
			)
			response.Error(w, http.StatusInternalServerError, "failed to update user")
			return
		}
		if count <= 1 {
			response.Error(w, http.StatusBadRequest, "cannot remove the last admin")
			return
		}
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
		h.log.Error("updating user",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
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

	currentUserID := middleware.GetUserID(r.Context())
	if id.Hex() == currentUserID {
		response.Error(w, http.StatusBadRequest, "cannot delete your own account")
		return
	}

	target, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("finding user for delete",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	if target.Role == model.RoleAdmin {
		count, err := h.repo.CountByRole(r.Context(), model.RoleAdmin)
		if err != nil {
			h.log.Error("counting admins",
				"error", err,
				"request_id", middleware.GetRequestID(r.Context()),
			)
			response.Error(w, http.StatusInternalServerError, "failed to delete user")
			return
		}
		if count <= 1 {
			response.Error(w, http.StatusBadRequest, "cannot delete the last admin")
			return
		}
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("deleting user",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
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

	if _, err := h.repo.FindByID(r.Context(), id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("finding user for password change",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to change password")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("hashing new password",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to change password")
		return
	}

	if err := h.repo.UpdatePassword(r.Context(), id, string(hashed)); err != nil {
		h.log.Error("updating password",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		response.Error(w, http.StatusInternalServerError, "failed to change password")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "password changed"})
}
