package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

	if err := seedAdminUser(ctx, db, adminUser, adminPass); err != nil {
		return fmt.Errorf("seeding admin user: %w", err)
	}
	if err := seedSiteSettings(ctx, db); err != nil {
		return fmt.Errorf("seeding site settings: %w", err)
	}
	if err := seedHero(ctx, db); err != nil {
		return fmt.Errorf("seeding hero: %w", err)
	}
	if err := seedAbout(ctx, db); err != nil {
		return fmt.Errorf("seeding about: %w", err)
	}
	if err := seedSkills(ctx, db); err != nil {
		return fmt.Errorf("seeding skills: %w", err)
	}
	if err := seedProjects(ctx, db); err != nil {
		return fmt.Errorf("seeding projects: %w", err)
	}
	if err := seedExperiences(ctx, db); err != nil {
		return fmt.Errorf("seeding experiences: %w", err)
	}
	if err := seedSocialLinks(ctx, db); err != nil {
		return fmt.Errorf("seeding social links: %w", err)
	}

	log.Info("database seeded successfully")
	return nil
}

func seedAdminUser(ctx context.Context, db *mongo.Database, username, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}
	now := time.Now()
	_, err = db.Collection("admin_users").InsertOne(ctx, model.AdminUser{
		Username:  username,
		Password:  string(hashed),
		Role:      model.RoleAdmin,
		CreatedAt: now,
		UpdatedAt: now,
	})
	return err
}

func seedSiteSettings(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection("site_settings").InsertOne(ctx, model.SiteSettings{
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

func seedHero(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection("hero").InsertOne(ctx, model.Hero{
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

func seedAbout(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection("about").InsertOne(ctx, model.About{
		Title: "Passionate about building great software",
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

func seedSkills(ctx context.Context, db *mongo.Database) error {
	now := time.Now()
	skills := []any{
		model.Skill{Name: "Vue.js", Icon: "logos:vue", Category: "frontend", SortOrder: 0, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "Nuxt", Icon: "logos:nuxt-icon", Category: "frontend", SortOrder: 1, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "TypeScript", Icon: "logos:typescript-icon", Category: "frontend", SortOrder: 2, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "TailwindCSS", Icon: "logos:tailwindcss-icon", Category: "frontend", SortOrder: 3, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "Go", Icon: "logos:go", Category: "backend", SortOrder: 4, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "Node.js", Icon: "logos:nodejs-icon-alt", Category: "backend", SortOrder: 5, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "MongoDB", Icon: "logos:mongodb-icon", Category: "backend", SortOrder: 6, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "Redis", Icon: "logos:redis", Category: "backend", SortOrder: 7, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "Docker", Icon: "logos:docker-icon", Category: "devops", SortOrder: 8, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "Kubernetes", Icon: "logos:kubernetes", Category: "devops", SortOrder: 9, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "GitHub Actions", Icon: "logos:github-actions", Category: "devops", SortOrder: 10, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "Git", Icon: "logos:git-icon", Category: "tools", SortOrder: 11, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "Figma", Icon: "logos:figma", Category: "tools", SortOrder: 12, CreatedAt: now, UpdatedAt: now},
		model.Skill{Name: "VS Code", Icon: "logos:visual-studio-code", Category: "tools", SortOrder: 13, CreatedAt: now, UpdatedAt: now},
	}
	_, err := db.Collection("skills").InsertMany(ctx, skills)
	return err
}

func seedProjects(ctx context.Context, db *mongo.Database) error {
	now := time.Now()
	projects := []any{
		model.Project{
			Title:       "E-Commerce Platform",
			Description: "A full-featured e-commerce platform with real-time inventory management and payment processing.",
			Tags:        []string{"Nuxt 3", "Go", "MongoDB", "Redis"},
			SortOrder:   0,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		model.Project{
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

func seedExperiences(ctx context.Context, db *mongo.Database) error {
	now := time.Now()
	experiences := []any{
		model.Experience{
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

func seedSocialLinks(ctx context.Context, db *mongo.Database) error {
	now := time.Now()
	links := []any{
		model.SocialLink{Name: "GitHub", URL: "https://github.com", Icon: "mdi:github", SortOrder: 0, CreatedAt: now, UpdatedAt: now},
		model.SocialLink{Name: "LinkedIn", URL: "https://linkedin.com", Icon: "mdi:linkedin", SortOrder: 1, CreatedAt: now, UpdatedAt: now},
		model.SocialLink{Name: "Twitter", URL: "https://twitter.com", Icon: "mdi:twitter", SortOrder: 2, CreatedAt: now, UpdatedAt: now},
		model.SocialLink{Name: "Email", URL: "mailto:hello@example.com", Icon: "mdi:email-outline", SortOrder: 3, CreatedAt: now, UpdatedAt: now},
	}
	_, err := db.Collection("social_links").InsertMany(ctx, links)
	return err
}
