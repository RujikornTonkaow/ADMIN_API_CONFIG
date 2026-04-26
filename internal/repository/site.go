package repository

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"portfolio-admin-api/internal/model"
)

type SiteRepository struct {
	col *mongo.Collection
}

func NewSiteRepository(db *mongo.Database) *SiteRepository {
	return &SiteRepository{col: db.Collection("sites")}
}

func normalizeDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return ""
	}

	if parsed, err := url.Parse(domain); err == nil && parsed.Host != "" {
		domain = parsed.Host
	}

	domain = strings.TrimPrefix(domain, "//")
	domain = strings.TrimSuffix(domain, "/")
	return strings.ToLower(domain)
}

func normalizeDomains(domains []string) []string {
	normalized := make([]string, 0, len(domains))
	seen := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		d := normalizeDomain(domain)
		if d == "" {
			continue
		}
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		normalized = append(normalized, d)
	}
	return normalized
}

func (r *SiteRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "domains", Value: 1}},
		},
	}
	_, err := r.col.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("creating site indexes: %w", err)
	}
	return nil
}

func (r *SiteRepository) Create(ctx context.Context, s model.Site) (model.Site, error) {
	now := time.Now()
	s.Domains = normalizeDomains(s.Domains)
	s.CreatedAt = now
	s.UpdatedAt = now

	result, err := r.col.InsertOne(ctx, s)
	if err != nil {
		return s, fmt.Errorf("inserting site: %w", err)
	}
	s.ID = result.InsertedID.(primitive.ObjectID)
	return s, nil
}

func (r *SiteRepository) FindByID(ctx context.Context, id primitive.ObjectID) (model.Site, error) {
	var site model.Site
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&site)
	if err != nil {
		return site, fmt.Errorf("finding site by id: %w", err)
	}
	return site, nil
}

func (r *SiteRepository) FindBySlug(ctx context.Context, slug string) (model.Site, error) {
	var site model.Site
	err := r.col.FindOne(ctx, bson.M{"slug": slug}).Decode(&site)
	if err != nil {
		return site, fmt.Errorf("finding site by slug: %w", err)
	}
	return site, nil
}

func (r *SiteRepository) FindByDomain(ctx context.Context, domain string) (model.Site, error) {
	var site model.Site
	err := r.col.FindOne(ctx, bson.M{"domains": normalizeDomain(domain)}).Decode(&site)
	if err != nil {
		return site, fmt.Errorf("finding site by domain: %w", err)
	}
	return site, nil
}

func (r *SiteRepository) ListByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model.Site, error) {
	cursor, err := r.col.Find(ctx, bson.M{"_id": bson.M{"$in": ids}},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("listing sites by ids: %w", err)
	}
	defer cursor.Close(ctx)

	var sites []model.Site
	if err := cursor.All(ctx, &sites); err != nil {
		return nil, fmt.Errorf("decoding sites: %w", err)
	}
	if sites == nil {
		sites = []model.Site{}
	}
	return sites, nil
}

func (r *SiteRepository) List(ctx context.Context) ([]model.Site, error) {
	cursor, err := r.col.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("listing sites: %w", err)
	}
	defer cursor.Close(ctx)

	var sites []model.Site
	if err := cursor.All(ctx, &sites); err != nil {
		return nil, fmt.Errorf("decoding sites: %w", err)
	}
	if sites == nil {
		sites = []model.Site{}
	}
	return sites, nil
}

func (r *SiteRepository) Update(ctx context.Context, id primitive.ObjectID, s model.UpdateSiteRequest) (model.Site, error) {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	domains := normalizeDomains(s.Domains)

	var result model.Site
	err := r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"name":       s.Name,
			"slug":       s.Slug,
			"domains":    domains,
			"updated_at": time.Now(),
		}},
		opts,
	).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("updating site: %w", err)
	}
	return result, nil
}

func (r *SiteRepository) ListAllDomains(ctx context.Context) ([]string, error) {
	cursor, err := r.col.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"domains": 1}))
	if err != nil {
		return nil, fmt.Errorf("listing all site domains: %w", err)
	}
	defer cursor.Close(ctx)

	var sites []model.Site
	if err := cursor.All(ctx, &sites); err != nil {
		return nil, fmt.Errorf("decoding site domains: %w", err)
	}

	var domains []string
	for _, s := range sites {
		domains = append(domains, normalizeDomains(s.Domains)...)
	}
	return domains, nil
}

func (r *SiteRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("deleting site: %w", err)
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
