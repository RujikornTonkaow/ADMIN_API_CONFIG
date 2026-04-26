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

type SiteMemberRepository struct {
	col *mongo.Collection
}

func NewSiteMemberRepository(db *mongo.Database) *SiteMemberRepository {
	return &SiteMemberRepository{col: db.Collection("site_members")}
}

func (r *SiteMemberRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "site_id", Value: 1}, {Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("creating site_member indexes: %w", err)
	}
	return nil
}

func (r *SiteMemberRepository) Create(ctx context.Context, m model.SiteMember) (model.SiteMember, error) {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now

	result, err := r.col.InsertOne(ctx, m)
	if err != nil {
		return m, fmt.Errorf("inserting site member: %w", err)
	}
	m.ID = result.InsertedID.(primitive.ObjectID)
	return m, nil
}

func (r *SiteMemberRepository) FindByID(ctx context.Context, id primitive.ObjectID) (model.SiteMember, error) {
	var member model.SiteMember
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&member)
	if err != nil {
		return member, fmt.Errorf("finding site member by id: %w", err)
	}
	return member, nil
}

func (r *SiteMemberRepository) FindBySiteAndUser(ctx context.Context, siteID, userID primitive.ObjectID) (model.SiteMember, error) {
	var member model.SiteMember
	err := r.col.FindOne(ctx, bson.M{"site_id": siteID, "user_id": userID}).Decode(&member)
	if err != nil {
		return member, fmt.Errorf("finding site member: %w", err)
	}
	return member, nil
}

func (r *SiteMemberRepository) ListBySite(ctx context.Context, siteID primitive.ObjectID) ([]model.SiteMember, error) {
	cursor, err := r.col.Find(ctx, bson.M{"site_id": siteID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("listing site members: %w", err)
	}
	defer cursor.Close(ctx)

	var members []model.SiteMember
	if err := cursor.All(ctx, &members); err != nil {
		return nil, fmt.Errorf("decoding site members: %w", err)
	}
	if members == nil {
		members = []model.SiteMember{}
	}
	return members, nil
}

func (r *SiteMemberRepository) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]model.SiteMember, error) {
	cursor, err := r.col.Find(ctx, bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("listing user memberships: %w", err)
	}
	defer cursor.Close(ctx)

	var members []model.SiteMember
	if err := cursor.All(ctx, &members); err != nil {
		return nil, fmt.Errorf("decoding user memberships: %w", err)
	}
	if members == nil {
		members = []model.SiteMember{}
	}
	return members, nil
}

func (r *SiteMemberRepository) UpdateRole(ctx context.Context, id primitive.ObjectID, role string) (model.SiteMember, error) {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result model.SiteMember
	err := r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"role":       role,
			"updated_at": time.Now(),
		}},
		opts,
	).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("updating site member role: %w", err)
	}
	return result, nil
}

func (r *SiteMemberRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("deleting site member: %w", err)
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *SiteMemberRepository) DeleteBySite(ctx context.Context, siteID primitive.ObjectID) error {
	_, err := r.col.DeleteMany(ctx, bson.M{"site_id": siteID})
	if err != nil {
		return fmt.Errorf("deleting site members by site: %w", err)
	}
	return nil
}

func (r *SiteMemberRepository) DeleteBySiteAndUser(ctx context.Context, siteID, userID primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"site_id": siteID, "user_id": userID})
	if err != nil {
		return fmt.Errorf("deleting site member by site and user: %w", err)
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *SiteMemberRepository) CountOwners(ctx context.Context, siteID primitive.ObjectID) (int64, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"site_id": siteID, "role": model.SiteRoleOwner})
	if err != nil {
		return 0, fmt.Errorf("counting site owners: %w", err)
	}
	return count, nil
}
