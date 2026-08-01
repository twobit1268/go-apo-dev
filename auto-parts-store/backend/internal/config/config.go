// Package config loads process configuration from environment variables,
// with sane local-dev defaults so `go run` works against docker-compose
// without any .env file.
package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	GCPProjectID string
}

func Load() Config {
	return Config{
		Port:         getenv("PORT", "8080"),
		DatabaseURL:  getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/autoparts?sslmode=disable"),
		GCPProjectID: getenv("GCP_PROJECT_ID", "auto-parts-local"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
