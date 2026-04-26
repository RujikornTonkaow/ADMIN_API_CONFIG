package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"portfolio-admin-api/internal/model"
)

func SeedIfEmpty(ctx context.Context, db *mongo.Database, adminUser, adminPass string, log *slog.Logger) error {
	count, err := db.Collection("admin_users").CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("checking admin_users count: %w", err)
	}
	if count > 0 {
		log.Info("database already seeded, skipping")
		return nil
	}

	log.Info("seeding database with initial data")

	adminID, err := seedAdminUser(ctx, db, adminUser, adminPass)
	if err != nil {
		return fmt.Errorf("seeding admin user: %w", err)
	}

	siteID, err := seedDefaultSite(ctx, db)
	if err != nil {
		return fmt.Errorf("seeding default site: %w", err)
	}

	if err := seedSiteMember(ctx, db, siteID, adminID); err != nil {
		return fmt.Errorf("seeding site member: %w", err)
	}
	if err := seedSiteSettings(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding site settings: %w", err)
	}
	if err := seedHero(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding hero: %w", err)
	}
	if err := seedAbout(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding about: %w", err)
	}
	if err := seedSkills(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding skills: %w", err)
	}
	if err := seedProjects(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding projects: %w", err)
	}
	if err := seedExperiences(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding experiences: %w", err)
	}
	if err := seedSocialLinks(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding social links: %w", err)
	}

	log.Info("database seeded successfully")
	return nil
}

func seedAdminUser(ctx context.Context, db *mongo.Database, username, password string) (primitive.ObjectID, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("hashing password: %w", err)
	}
	now := time.Now()
	result, err := db.Collection("admin_users").InsertOne(ctx, model.AdminUser{
		Username:  username,
		Password:  string(hashed),
		Role:      model.RoleSuperAdmin,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func seedDefaultSite(ctx context.Context, db *mongo.Database) (primitive.ObjectID, error) {
	now := time.Now()
	result, err := db.Collection("sites").InsertOne(ctx, model.Site{
		Name:      "My Portfolio",
		Slug:      "default",
		Type:      model.SiteTypePortfolio,
		Domains:   []string{"localhost:3000"},
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("inserting default site: %w", err)
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func seedSiteMember(ctx context.Context, db *mongo.Database, siteID, userID primitive.ObjectID) error {
	now := time.Now()
	_, err := db.Collection("site_members").InsertOne(ctx, model.SiteMember{
		SiteID:    siteID,
		UserID:    userID,
		Role:      model.SiteRoleViewer,
		CreatedAt: now,
		UpdatedAt: now,
	})
	return err
}

func SeedPortfolioContent(ctx context.Context, db *mongo.Database, siteID primitive.ObjectID) error {
	if err := seedSiteSettings(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding site settings: %w", err)
	}
	if err := seedHero(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding hero: %w", err)
	}
	if err := seedAbout(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding about: %w", err)
	}
	if err := seedSkills(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding skills: %w", err)
	}
	if err := seedProjects(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding projects: %w", err)
	}
	if err := seedExperiences(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding experiences: %w", err)
	}
	if err := seedSocialLinks(ctx, db, siteID); err != nil {
		return fmt.Errorf("seeding social links: %w", err)
	}
	return nil
}

func EnsurePortfolioContent(ctx context.Context, db *mongo.Database, siteID primitive.ObjectID) error {
	count, err := db.Collection("site_settings").CountDocuments(ctx, bson.M{"site_id": siteID})
	if err != nil {
		return fmt.Errorf("checking site settings: %w", err)
	}
	if count > 0 {
		return nil
	}
	return SeedPortfolioContent(ctx, db, siteID)
}

func seedSiteSettings(ctx context.Context, db *mongo.Database, siteID primitive.ObjectID) error {
	_, err := db.Collection("site_settings").InsertOne(ctx, model.SiteSettings{
		SiteID:          siteID,
		SiteTitle:       "Portfolio",
		PageTitle:       "Portfolio | Full-Stack Developer",
		MetaDescription: "Full-Stack Developer portfolio showcasing projects, skills, and experience",
		FooterTagline:   "Crafting digital experiences",
		DefaultTheme:    "midnight",
		ProfileImage:    "/images/profile.jpg",
		UpdatedAt:       time.Now(),
	})
	return err
}

func seedHero(ctx context.Context, db *mongo.Database, siteID primitive.ObjectID) error {
	_, err := db.Collection("hero").InsertOne(ctx, model.Hero{
		SiteID:           siteID,
		Greeting:         "Hello, I'm",
		FullName:         "Puvakorn Pannasirichard",
		Subtitle:         "Full-Stack Developer crafting performant, scalable web applications with modern technologies",
		CTAPrimaryText:   "View My Work",
		CTAPrimaryLink:   "#projects",
		CTASecondaryText: "Get in Touch",
		CTASecondaryLink: "#contact",
		UpdatedAt:        time.Now(),
	})
	return err
}

func seedAbout(ctx context.Context, db *mongo.Database, siteID primitive.ObjectID) error {
	_, err := db.Collection("about").InsertOne(ctx, model.About{
		SiteID: siteID,
		Title:  "Passionate about building great software",
		BioParagraphs: []string{
			"I'm a Full-Stack Developer with a passion for creating elegant, efficient, and user-friendly web applications. With expertise in both frontend and backend technologies, I bring ideas to life from concept to deployment.",
			"When I'm not coding, you'll find me exploring new technologies, contributing to open-source projects, and sharing knowledge with the developer community.",
		},
		PersonalityTags: []string{"Problem Solver", "Team Player", "Continuous Learner"},
		Stats: []model.Stat{
			{Value: "5+", Label: "Years Experience"},
			{Value: "30+", Label: "Projects Completed"},
			{Value: "15+", Label: "Happy Clients"},
			{Value: "99%", Label: "Client Satisfaction"},
		},
		UpdatedAt: time.Now(),
	})
	return err
}

func seedSkills(ctx context.Context, db *mongo.Database, siteID primitive.ObjectID) error {
	now := time.Now()
	skills := []any{
		model.Skill{SiteID: siteID, Name: "Vue.js", Icon: "logos:vue", Category: "frontend", SortOrder: 0, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "Nuxt", Icon: "logos:nuxt-icon", Category: "frontend", SortOrder: 1, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "TypeScript", Icon: "logos:typescript-icon", Category: "frontend", SortOrder: 2, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "TailwindCSS", Icon: "logos:tailwindcss-icon", Category: "frontend", SortOrder: 3, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "Go", Icon: "logos:go", Category: "backend", SortOrder: 4, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "Node.js", Icon: "logos:nodejs-icon-alt", Category: "backend", SortOrder: 5, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "MongoDB", Icon: "logos:mongodb-icon", Category: "backend", SortOrder: 6, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "Redis", Icon: "logos:redis", Category: "backend", SortOrder: 7, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "Docker", Icon: "logos:docker-icon", Category: "devops", SortOrder: 8, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "Kubernetes", Icon: "logos:kubernetes", Category: "devops", SortOrder: 9, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "GitHub Actions", Icon: "logos:github-actions", Category: "devops", SortOrder: 10, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "Git", Icon: "logos:git-icon", Category: "tools", SortOrder: 11, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "Figma", Icon: "logos:figma", Category: "tools", SortOrder: 12, CreatedAt: now, UpdatedAt: now},
		model.Skill{SiteID: siteID, Name: "VS Code", Icon: "logos:visual-studio-code", Category: "tools", SortOrder: 13, CreatedAt: now, UpdatedAt: now},
	}
	_, err := db.Collection("skills").InsertMany(ctx, skills)
	return err
}

func seedProjects(ctx context.Context, db *mongo.Database, siteID primitive.ObjectID) error {
	now := time.Now()
	projects := []any{
		model.Project{
			SiteID:      siteID,
			Title:       "E-Commerce Platform",
			Description: "A full-featured e-commerce platform with real-time inventory management and payment processing.",
			Tags:        []string{"Nuxt 3", "Go", "MongoDB", "Redis"},
			SortOrder:   0,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		model.Project{
			SiteID:      siteID,
			Title:       "Task Management App",
			Description: "A collaborative task management application with real-time updates and team features.",
			Tags:        []string{"Vue 3", "Node.js", "MongoDB", "WebSocket"},
			SortOrder:   1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	_, err := db.Collection("projects").InsertMany(ctx, projects)
	return err
}

func seedExperiences(ctx context.Context, db *mongo.Database, siteID primitive.ObjectID) error {
	now := time.Now()
	experiences := []any{
		model.Experience{
			SiteID:      siteID,
			Role:        "Senior Full-Stack Developer",
			Company:     "Tech Company",
			Period:      "2024 - Present",
			Description: "Leading development of scalable web applications and microservices architecture.",
			Highlights: []string{
				"Architected and built microservices handling 10K+ requests per second",
				"Reduced deployment time by 60% through CI/CD pipeline optimization",
				"Mentored junior developers and conducted code reviews",
			},
			SortOrder: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		model.Experience{
			SiteID:      siteID,
			Role:        "Full-Stack Developer",
			Company:     "Digital Agency",
			Period:      "2021 - 2024",
			Description: "Developed custom web applications for diverse clients across multiple industries.",
			Highlights: []string{
				"Delivered 20+ client projects on time and within budget",
				"Built reusable component library reducing development time by 40%",
				"Implemented responsive designs achieving 95+ Lighthouse scores",
			},
			SortOrder: 1,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	_, err := db.Collection("experiences").InsertMany(ctx, experiences)
	return err
}

func seedSocialLinks(ctx context.Context, db *mongo.Database, siteID primitive.ObjectID) error {
	now := time.Now()
	links := []any{
		model.SocialLink{SiteID: siteID, Name: "GitHub", URL: "https://github.com", Icon: "mdi:github", SortOrder: 0, CreatedAt: now, UpdatedAt: now},
		model.SocialLink{SiteID: siteID, Name: "LinkedIn", URL: "https://linkedin.com", Icon: "mdi:linkedin", SortOrder: 1, CreatedAt: now, UpdatedAt: now},
		model.SocialLink{SiteID: siteID, Name: "Twitter", URL: "https://twitter.com", Icon: "mdi:twitter", SortOrder: 2, CreatedAt: now, UpdatedAt: now},
		model.SocialLink{SiteID: siteID, Name: "Email", URL: "mailto:hello@example.com", Icon: "mdi:email-outline", SortOrder: 3, CreatedAt: now, UpdatedAt: now},
	}
	_, err := db.Collection("social_links").InsertMany(ctx, links)
	return err
}
