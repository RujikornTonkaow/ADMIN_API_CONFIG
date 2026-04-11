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

type SiteSettingsRepository struct {
	col *mongo.Collection
}

func NewSiteSettingsRepository(db *mongo.Database) *SiteSettingsRepository {
	return &SiteSettingsRepository{col: db.Collection("site_settings")}
}

func (r *SiteSettingsRepository) Get(ctx context.Context) (model.SiteSettings, error) {
	var settings model.SiteSettings
	err := r.col.FindOne(ctx, bson.M{}).Decode(&settings)
	if err != nil {
		return settings, fmt.Errorf("finding site settings: %w", err)
	}
	return settings, nil
}

func (r *SiteSettingsRepository) Upsert(ctx context.Context, s model.SiteSettings) (model.SiteSettings, error) {
	s.UpdatedAt = time.Now()
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result model.SiteSettings
	err := r.col.FindOneAndUpdate(
		ctx,
		bson.M{},
		bson.M{"$set": s},
		opts,
	).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("upserting site settings: %w", err)
	}
	return result, nil
}
