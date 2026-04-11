package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"portfolio-admin-api/internal/model"
)

type AdminUserRepository struct {
	col *mongo.Collection
}

func NewAdminUserRepository(db *mongo.Database) *AdminUserRepository {
	return &AdminUserRepository{col: db.Collection("admin_users")}
}

func (r *AdminUserRepository) FindByUsername(ctx context.Context, username string) (model.AdminUser, error) {
	var user model.AdminUser
	err := r.col.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return user, fmt.Errorf("finding admin user by username: %w", err)
	}
	return user, nil
}
