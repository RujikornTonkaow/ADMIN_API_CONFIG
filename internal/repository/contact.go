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

type ContactRepository struct {
	col *mongo.Collection
}

func NewContactRepository(db *mongo.Database) *ContactRepository {
	return &ContactRepository{col: db.Collection("contact_messages")}
}

func (r *ContactRepository) List(ctx context.Context, siteID primitive.ObjectID) ([]model.ContactMessage, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.col.Find(ctx, bson.M{"site_id": siteID}, opts)
	if err != nil {
		return nil, fmt.Errorf("listing contact messages: %w", err)
	}
	defer cursor.Close(ctx)

	var messages []model.ContactMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, fmt.Errorf("decoding contact messages: %w", err)
	}
	if messages == nil {
		messages = []model.ContactMessage{}
	}
	return messages, nil
}

func (r *ContactRepository) GetByID(ctx context.Context, siteID, id primitive.ObjectID) (model.ContactMessage, error) {
	var msg model.ContactMessage
	err := r.col.FindOne(ctx, bson.M{"_id": id, "site_id": siteID}).Decode(&msg)
	if err != nil {
		return msg, fmt.Errorf("finding contact message: %w", err)
	}
	return msg, nil
}

func (r *ContactRepository) Create(ctx context.Context, siteID primitive.ObjectID, c model.ContactMessage) (model.ContactMessage, error) {
	c.SiteID = siteID
	c.CreatedAt = time.Now()
	c.IsRead = false

	result, err := r.col.InsertOne(ctx, c)
	if err != nil {
		return c, fmt.Errorf("inserting contact message: %w", err)
	}
	c.ID = result.InsertedID.(primitive.ObjectID)
	return c, nil
}

func (r *ContactRepository) MarkAsRead(ctx context.Context, siteID, id primitive.ObjectID) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id, "site_id": siteID},
		bson.M{"$set": bson.M{"is_read": true}},
	)
	if err != nil {
		return fmt.Errorf("marking contact message as read: %w", err)
	}
	return nil
}

func (r *ContactRepository) Delete(ctx context.Context, siteID, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "site_id": siteID})
	if err != nil {
		return fmt.Errorf("deleting contact message: %w", err)
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *ContactRepository) CountUnread(ctx context.Context, siteID primitive.ObjectID) (int64, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"site_id": siteID, "is_read": false})
	if err != nil {
		return 0, fmt.Errorf("counting unread messages: %w", err)
	}
	return count, nil
}
