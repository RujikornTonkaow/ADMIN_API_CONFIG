package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"portfolio-admin-api/internal/config"
	"portfolio-admin-api/internal/database"
	"portfolio-admin-api/internal/middleware"
	"portfolio-admin-api/internal/repository"
	"portfolio-admin-api/internal/router"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	cfg := config.Load()

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Error("creating upload directory", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	db, disconnect, err := database.Connect(ctx, cfg.MongoURI, cfg.MongoDB, log)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer disconnect()

	if cfg.ResetDatabaseOnStart {
		log.Warn("resetting database on start", "database", cfg.MongoDB)
		if err := db.Drop(ctx); err != nil {
			log.Error("failed to reset database", "error", err)
			os.Exit(1)
		}
	}

	adminUserRepo := repository.NewAdminUserRepository(db)
	siteRepo := repository.NewSiteRepository(db)
	siteMemberRepo := repository.NewSiteMemberRepository(db)

	if err := adminUserRepo.EnsureIndexes(ctx); err != nil {
		log.Error("failed to ensure admin_users indexes", "error", err)
		os.Exit(1)
	}
	if err := siteRepo.EnsureIndexes(ctx); err != nil {
		log.Error("failed to ensure sites indexes", "error", err)
		os.Exit(1)
	}
	if err := siteMemberRepo.EnsureIndexes(ctx); err != nil {
		log.Error("failed to ensure site_members indexes", "error", err)
		os.Exit(1)
	}

	if err := database.SeedIfEmpty(ctx, db, cfg.AdminUsername, cfg.AdminPassword, log); err != nil {
		log.Error("failed to seed database", "error", err)
		os.Exit(1)
	}

	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	domainCache := middleware.NewDomainCache(siteRepo, cfg.AllowedOrigins, log)
	domainCache.StartAutoRefresh(appCtx, 5*time.Minute)

	siteSettingsRepo := repository.NewSiteSettingsRepository(db)
	heroRepo := repository.NewHeroRepository(db)
	aboutRepo := repository.NewAboutRepository(db)
	skillRepo := repository.NewSkillRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	experienceRepo := repository.NewExperienceRepository(db)
	socialLinkRepo := repository.NewSocialLinkRepository(db)
	contactRepo := repository.NewContactRepository(db)

	handler := router.New(&router.Config{
		JWTSecret:   cfg.JWTSecret,
		UploadDir:   cfg.UploadDir,
		MaxUploadMB: cfg.MaxUploadSizeMB,
		DomainCache: domainCache,
		Log:         log,
		DB:          db,
		SiteSettingsRepo: siteSettingsRepo,
		HeroRepo:         heroRepo,
		AboutRepo:        aboutRepo,
		SkillRepo:        skillRepo,
		ProjectRepo:      projectRepo,
		ExperienceRepo:   experienceRepo,
		SocialLinkRepo:   socialLinkRepo,
		ContactRepo:      contactRepo,
		AdminUserRepo:    adminUserRepo,
		SiteRepo:         siteRepo,
		SiteMemberRepo:   siteMemberRepo,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		log.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	log.Info("shutdown signal received", "signal", sig.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped gracefully")
}
