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

type SocialLinkRepository struct {
	col *mongo.Collection
}

func NewSocialLinkRepository(db *mongo.Database) *SocialLinkRepository {
	return &SocialLinkRepository{col: db.Collection("social_links")}
}

func (r *SocialLinkRepository) List(ctx context.Context) ([]model.SocialLink, error) {
	opts := options.Find().SetSort(bson.D{{Key: "sort_order", Value: 1}})
	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("listing social links: %w", err)
	}
	defer cursor.Close(ctx)

	var links []model.SocialLink
	if err := cursor.All(ctx, &links); err != nil {
		return nil, fmt.Errorf("decoding social links: %w", err)
	}
	if links == nil {
		links = []model.SocialLink{}
	}
	return links, nil
}

func (r *SocialLinkRepository) Create(ctx context.Context, s model.SocialLink) (model.SocialLink, error) {
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now

	count, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return s, fmt.Errorf("counting social links for sort order: %w", err)
	}
	s.SortOrder = int(count)

	result, err := r.col.InsertOne(ctx, s)
	if err != nil {
		return s, fmt.Errorf("inserting social link: %w", err)
	}
	s.ID = result.InsertedID.(primitive.ObjectID)
	return s, nil
}

func (r *SocialLinkRepository) Update(ctx context.Context, id primitive.ObjectID, s model.SocialLink) (model.SocialLink, error) {
	s.UpdatedAt = time.Now()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result model.SocialLink
	err := r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"name":       s.Name,
			"url":        s.URL,
			"icon":       s.Icon,
			"updated_at": s.UpdatedAt,
		}},
		opts,
	).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("updating social link: %w", err)
	}
	return result, nil
}

func (r *SocialLinkRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("deleting social link: %w", err)
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *SocialLinkRepository) Reorder(ctx context.Context, ids []primitive.ObjectID) error {
	for i, id := range ids {
		_, err := r.col.UpdateOne(ctx,
			bson.M{"_id": id},
			bson.M{"$set": bson.M{"sort_order": i}},
		)
		if err != nil {
			return fmt.Errorf("reordering social link %s: %w", id.Hex(), err)
		}
	}
	return nil
}
