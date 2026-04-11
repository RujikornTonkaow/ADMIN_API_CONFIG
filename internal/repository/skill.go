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

type SkillRepository struct {
	col *mongo.Collection
}

func NewSkillRepository(db *mongo.Database) *SkillRepository {
	return &SkillRepository{col: db.Collection("skills")}
}

func (r *SkillRepository) List(ctx context.Context) ([]model.Skill, error) {
	opts := options.Find().SetSort(bson.D{{Key: "sort_order", Value: 1}})
	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("listing skills: %w", err)
	}
	defer cursor.Close(ctx)

	var skills []model.Skill
	if err := cursor.All(ctx, &skills); err != nil {
		return nil, fmt.Errorf("decoding skills: %w", err)
	}
	if skills == nil {
		skills = []model.Skill{}
	}
	return skills, nil
}

func (r *SkillRepository) Create(ctx context.Context, s model.Skill) (model.Skill, error) {
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now

	count, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return s, fmt.Errorf("counting skills for sort order: %w", err)
	}
	s.SortOrder = int(count)

	result, err := r.col.InsertOne(ctx, s)
	if err != nil {
		return s, fmt.Errorf("inserting skill: %w", err)
	}
	s.ID = result.InsertedID.(primitive.ObjectID)
	return s, nil
}

func (r *SkillRepository) Update(ctx context.Context, id primitive.ObjectID, s model.Skill) (model.Skill, error) {
	s.UpdatedAt = time.Now()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result model.Skill
	err := r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"name":       s.Name,
			"icon":       s.Icon,
			"category":   s.Category,
			"updated_at": s.UpdatedAt,
		}},
		opts,
	).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("updating skill: %w", err)
	}
	return result, nil
}

func (r *SkillRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("deleting skill: %w", err)
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
