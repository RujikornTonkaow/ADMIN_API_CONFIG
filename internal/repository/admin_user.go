package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"portfolio-admin-api/internal/model"
)

type AdminUserRepository struct {
	col *mongo.Collection
}

func NewAdminUserRepository(db *mongo.Database) *AdminUserRepository {
	return &AdminUserRepository{col: db.Collection("admin_users")}
}

func (r *AdminUserRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("creating unique index on username: %w", err)
	}
	return nil
}

func (r *AdminUserRepository) FindByUsername(ctx context.Context, username string) (model.AdminUser, error) {
	var user model.AdminUser
	err := r.col.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return user, fmt.Errorf("finding admin user by username: %w", err)
	}
	return user, nil
}

func (r *AdminUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (model.AdminUser, error) {
	var user model.AdminUser
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return user, fmt.Errorf("finding admin user by id: %w", err)
	}
	return user, nil
}

func (r *AdminUserRepository) List(ctx context.Context) ([]model.AdminUser, error) {
	cursor, err := r.col.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("listing admin users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []model.AdminUser
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("decoding admin users: %w", err)
	}
	if users == nil {
		users = []model.AdminUser{}
	}
	return users, nil
}

func (r *AdminUserRepository) Create(ctx context.Context, user model.AdminUser) (model.AdminUser, error) {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	result, err := r.col.InsertOne(ctx, user)
	if err != nil {
		return user, fmt.Errorf("creating admin user: %w", err)
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

func (r *AdminUserRepository) Update(ctx context.Context, id primitive.ObjectID, username, role string) (model.AdminUser, error) {
	update := bson.M{
		"$set": bson.M{
			"username":   username,
			"role":       role,
			"updated_at": time.Now(),
		},
	}

	var user model.AdminUser
	err := r.col.FindOneAndUpdate(
		ctx, bson.M{"_id": id}, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&user)
	if err != nil {
		return user, fmt.Errorf("updating admin user: %w", err)
	}
	return user, nil
}

func (r *AdminUserRepository) UpdatePassword(ctx context.Context, id primitive.ObjectID, hashedPassword string) error {
	update := bson.M{
		"$set": bson.M{
			"password":   hashedPassword,
			"updated_at": time.Now(),
		},
	}
	result, err := r.col.UpdateByID(ctx, id, update)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *AdminUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("deleting admin user: %w", err)
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *AdminUserRepository) CountByRole(ctx context.Context, role string) (int64, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"role": role})
	if err != nil {
		return 0, fmt.Errorf("counting users by role: %w", err)
	}
	return count, nil
}
