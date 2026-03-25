package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTP          HTTPConfig
	Database      DatabaseConfig
	Migrations    MigrationsConfig
	JWT           JWTConfig
	ReadinessPing ReadinessConfig
}

type HTTPConfig struct {
	Port string
}

type DatabaseConfig struct {
	URL string
}

type MigrationsConfig struct {
	Path string
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

type ReadinessConfig struct {
	DBPingTimeout time.Duration
}

func LoadFromEnv() (Config, error) {
	jwtTTLHours, err := mustPositiveInt("JWT_TTL_HOURS", "24")
	if err != nil {
		return Config{}, err
	}
	dbPingSeconds, err := mustPositiveInt("DB_PING_TIMEOUT_SECONDS", "5")
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://booking:booking@localhost:5432/booking?sslmode=disable"),
		},
		Migrations: MigrationsConfig{
			Path: getEnv("MIGRATIONS_PATH", "migrations"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "dev-secret"),
			TTL:    time.Duration(jwtTTLHours) * time.Hour,
		},
		ReadinessPing: ReadinessConfig{
			DBPingTimeout: time.Duration(dbPingSeconds) * time.Second,
		},
	}

	if strings.TrimSpace(cfg.Database.URL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is empty")
	}
	if strings.TrimSpace(cfg.HTTP.Port) == "" {
		return Config{}, fmt.Errorf("HTTP_PORT is empty")
	}
	if strings.TrimSpace(cfg.JWT.Secret) == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is empty")
	}
	if strings.TrimSpace(cfg.Migrations.Path) == "" {
		return Config{}, fmt.Errorf("MIGRATIONS_PATH is empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func mustPositiveInt(key string, fallback string) (int, error) {
	raw := getEnv(key, fallback)
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid %s: %q", key, raw)
	}
	return value, nil
}
