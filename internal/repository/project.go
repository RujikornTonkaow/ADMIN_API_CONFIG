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

type ProjectRepository struct {
	col *mongo.Collection
}

func NewProjectRepository(db *mongo.Database) *ProjectRepository {
	return &ProjectRepository{col: db.Collection("projects")}
}

func (r *ProjectRepository) List(ctx context.Context, siteID primitive.ObjectID) ([]model.Project, error) {
	opts := options.Find().SetSort(bson.D{{Key: "sort_order", Value: 1}})
	cursor, err := r.col.Find(ctx, bson.M{"site_id": siteID}, opts)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	defer cursor.Close(ctx)

	var projects []model.Project
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, fmt.Errorf("decoding projects: %w", err)
	}
	if projects == nil {
		projects = []model.Project{}
	}
	return projects, nil
}

func (r *ProjectRepository) Create(ctx context.Context, siteID primitive.ObjectID, p model.Project) (model.Project, error) {
	now := time.Now()
	p.SiteID = siteID
	p.CreatedAt = now
	p.UpdatedAt = now

	count, err := r.col.CountDocuments(ctx, bson.M{"site_id": siteID})
	if err != nil {
		return p, fmt.Errorf("counting projects for sort order: %w", err)
	}
	p.SortOrder = int(count)

	result, err := r.col.InsertOne(ctx, p)
	if err != nil {
		return p, fmt.Errorf("inserting project: %w", err)
	}
	p.ID = result.InsertedID.(primitive.ObjectID)
	return p, nil
}

func (r *ProjectRepository) Update(ctx context.Context, siteID, id primitive.ObjectID, p model.Project) (model.Project, error) {
	p.UpdatedAt = time.Now()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result model.Project
	err := r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id, "site_id": siteID},
		bson.M{"$set": bson.M{
			"title":       p.Title,
			"description": p.Description,
			"tags":        p.Tags,
			"image":       p.Image,
			"live_url":    p.LiveURL,
			"source_url":  p.SourceURL,
			"updated_at":  p.UpdatedAt,
		}},
		opts,
	).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("updating project: %w", err)
	}
	return result, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, siteID, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "site_id": siteID})
	if err != nil {
		return fmt.Errorf("deleting project: %w", err)
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *ProjectRepository) Reorder(ctx context.Context, siteID primitive.ObjectID, ids []primitive.ObjectID) error {
	for i, id := range ids {
		_, err := r.col.UpdateOne(ctx,
			bson.M{"_id": id, "site_id": siteID},
			bson.M{"$set": bson.M{"sort_order": i}},
		)
		if err != nil {
			return fmt.Errorf("reordering project %s: %w", id.Hex(), err)
		}
	}
	return nil
}
