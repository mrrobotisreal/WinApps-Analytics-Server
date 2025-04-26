package config

import (
	"os"
	"time"
)

type Config struct {
	PostgresURL   string
	RedisURL      string
	Port          string
	FlushInterval time.Duration
	SiteSecret    string
}

func Load() Config {
	cfg := Config{
		PostgresURL:   getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/analytics?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		Port:          getEnv("PORT", "17177"),
		FlushInterval: 5 * time.Second,
		SiteSecret:    getEnv("SITE_SECRET", "ahhh-im-exposed"),
	}
	return cfg
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
