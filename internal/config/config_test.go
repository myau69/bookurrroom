package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadFromEnvDefaults(t *testing.T) {
	reset := snapshotEnv(t, []string{
		"HTTP_PORT",
		"DATABASE_URL",
		"MIGRATIONS_PATH",
		"JWT_SECRET",
		"JWT_TTL_HOURS",
		"DB_PING_TIMEOUT_SECONDS",
	})
	defer reset()

	clearEnv("HTTP_PORT", "DATABASE_URL", "MIGRATIONS_PATH", "JWT_SECRET", "JWT_TTL_HOURS", "DB_PING_TIMEOUT_SECONDS")

	cfg, err := LoadFromEnv()
	require.NoError(t, err)
	require.Equal(t, "8080", cfg.HTTP.Port)
	require.Equal(t, "dev-secret", cfg.JWT.Secret)
	require.Equal(t, 24*time.Hour, cfg.JWT.TTL)
}

func TestLoadFromEnvCustomValues(t *testing.T) {
	reset := snapshotEnv(t, []string{
		"HTTP_PORT",
		"DATABASE_URL",
		"MIGRATIONS_PATH",
		"JWT_SECRET",
		"JWT_TTL_HOURS",
		"DB_PING_TIMEOUT_SECONDS",
	})
	defer reset()

	require.NoError(t, os.Setenv("HTTP_PORT", "18080"))
	require.NoError(t, os.Setenv("DATABASE_URL", "postgres://x"))
	require.NoError(t, os.Setenv("MIGRATIONS_PATH", "/tmp/migrations"))
	require.NoError(t, os.Setenv("JWT_SECRET", "secret-123"))
	require.NoError(t, os.Setenv("JWT_TTL_HOURS", "48"))
	require.NoError(t, os.Setenv("DB_PING_TIMEOUT_SECONDS", "9"))

	cfg, err := LoadFromEnv()
	require.NoError(t, err)
	require.Equal(t, "18080", cfg.HTTP.Port)
	require.Equal(t, "postgres://x", cfg.Database.URL)
	require.Equal(t, "/tmp/migrations", cfg.Migrations.Path)
	require.Equal(t, "secret-123", cfg.JWT.Secret)
	require.Equal(t, 48*time.Hour, cfg.JWT.TTL)
}

func TestLoadFromEnvInvalidValue(t *testing.T) {
	reset := snapshotEnv(t, []string{"JWT_TTL_HOURS"})
	defer reset()

	require.NoError(t, os.Setenv("JWT_TTL_HOURS", "abc"))

	_, err := LoadFromEnv()
	require.Error(t, err)
}

func snapshotEnv(t *testing.T, keys []string) func() {
	t.Helper()
	values := make(map[string]*string, len(keys))
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if ok {
			v := value
			values[key] = &v
		} else {
			values[key] = nil
		}
	}

	return func() {
		for _, key := range keys {
			if values[key] == nil {
				_ = os.Unsetenv(key)
				continue
			}
			_ = os.Setenv(key, *values[key])
		}
	}
}

func clearEnv(keys ...string) {
	for _, key := range keys {
		_ = os.Unsetenv(key)
	}
}
