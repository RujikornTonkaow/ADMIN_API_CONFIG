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

type ExperienceRepository struct {
	col *mongo.Collection
}

func NewExperienceRepository(db *mongo.Database) *ExperienceRepository {
	return &ExperienceRepository{col: db.Collection("experiences")}
}

func (r *ExperienceRepository) List(ctx context.Context) ([]model.Experience, error) {
	opts := options.Find().SetSort(bson.D{{Key: "sort_order", Value: 1}})
	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("listing experiences: %w", err)
	}
	defer cursor.Close(ctx)

	var experiences []model.Experience
	if err := cursor.All(ctx, &experiences); err != nil {
		return nil, fmt.Errorf("decoding experiences: %w", err)
	}
	if experiences == nil {
		experiences = []model.Experience{}
	}
	return experiences, nil
}

func (r *ExperienceRepository) Create(ctx context.Context, e model.Experience) (model.Experience, error) {
	now := time.Now()
	e.CreatedAt = now
	e.UpdatedAt = now

	count, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return e, fmt.Errorf("counting experiences for sort order: %w", err)
	}
	e.SortOrder = int(count)

	result, err := r.col.InsertOne(ctx, e)
	if err != nil {
		return e, fmt.Errorf("inserting experience: %w", err)
	}
	e.ID = result.InsertedID.(primitive.ObjectID)
	return e, nil
}

func (r *ExperienceRepository) Update(ctx context.Context, id primitive.ObjectID, e model.Experience) (model.Experience, error) {
	e.UpdatedAt = time.Now()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result model.Experience
	err := r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"role":        e.Role,
			"company":     e.Company,
			"period":      e.Period,
			"description": e.Description,
			"highlights":  e.Highlights,
			"updated_at":  e.UpdatedAt,
		}},
		opts,
	).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("updating experience: %w", err)
	}
	return result, nil
}

func (r *ExperienceRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("deleting experience: %w", err)
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *ExperienceRepository) Reorder(ctx context.Context, ids []primitive.ObjectID) error {
	for i, id := range ids {
		_, err := r.col.UpdateOne(ctx,
			bson.M{"_id": id},
			bson.M{"$set": bson.M{"sort_order": i}},
		)
		if err != nil {
			return fmt.Errorf("reordering experience %s: %w", id.Hex(), err)
		}
	}
	return nil
}
