package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                 string
	MongoURI             string
	MongoDB              string
	JWTSecret            string
	AdminUsername        string
	AdminPassword        string
	UploadDir            string
	MaxUploadSizeMB      int64
	AllowedOrigins       string
	ResetDatabaseOnStart bool
}

func Load() *Config {
	return &Config{
		Port:                 getEnv("PORT", "8080"),
		MongoURI:             getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:              getEnv("MONGO_DB", "portfolio_admin"),
		JWTSecret:            getEnv("JWT_SECRET", "change-me-in-production"),
		AdminUsername:        getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:        getEnv("ADMIN_PASSWORD", "changeme123"),
		UploadDir:            getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSizeMB:      getEnvInt("MAX_UPLOAD_SIZE_MB", 10),
		AllowedOrigins:       getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		ResetDatabaseOnStart: getEnvBool("RESET_DATABASE_ON_START", false),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
