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

type HeroRepository struct {
	col *mongo.Collection
}

func NewHeroRepository(db *mongo.Database) *HeroRepository {
	return &HeroRepository{col: db.Collection("hero")}
}

func (r *HeroRepository) Get(ctx context.Context) (model.Hero, error) {
	var hero model.Hero
	err := r.col.FindOne(ctx, bson.M{}).Decode(&hero)
	if err != nil {
		return hero, fmt.Errorf("finding hero: %w", err)
	}
	return hero, nil
}

func (r *HeroRepository) Upsert(ctx context.Context, h model.Hero) (model.Hero, error) {
	h.UpdatedAt = time.Now()
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result model.Hero
	err := r.col.FindOneAndUpdate(
		ctx,
		bson.M{},
		bson.M{"$set": h},
		opts,
	).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("upserting hero: %w", err)
	}
	return result, nil
}
