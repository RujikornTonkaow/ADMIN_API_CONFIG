package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"portfolio-admin-api/internal/model"
)

type AboutRepository struct {
	col *mongo.Collection
}

func NewAboutRepository(db *mongo.Database) *AboutRepository {
	return &AboutRepository{col: db.Collection("about")}
}

func (r *AboutRepository) Get(ctx context.Context) (model.About, error) {
	var about model.About
	err := r.col.FindOne(ctx, bson.M{}).Decode(&about)
	if err != nil {
		return about, fmt.Errorf("finding about: %w", err)
	}
	return about, nil
}

func (r *AboutRepository) Upsert(ctx context.Context, a model.About) (model.About, error) {
	a.UpdatedAt = time.Now()
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result model.About
	err := r.col.FindOneAndUpdate(
		ctx,
		bson.M{},
		bson.M{"$set": a},
		opts,
	).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("upserting about: %w", err)
	}
	return result, nil
}
