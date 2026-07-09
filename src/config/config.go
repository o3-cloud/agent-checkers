// Package config loads runtime configuration.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/stackable-specs/agent-checkers/internal/app/store"
)

// StoreConfig configures the persistence backend.
type StoreConfig struct {
	Type          string // "memory" | "redis"
	RedisURL      string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisPoolSize int
}

// LoadStoreConfig reads store configuration from environment variables.
func LoadStoreConfig() StoreConfig {
	cfg := StoreConfig{
		Type:          strings.ToLower(strings.TrimSpace(os.Getenv("STORE_TYPE"))),
		RedisURL:      strings.TrimSpace(os.Getenv("REDIS_URL")),
		RedisAddr:     envString("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       envInt("REDIS_DB", 0),
		RedisPoolSize: envInt("REDIS_POOL_SIZE", 10),
	}
	if cfg.RedisURL != "" {
		cfg.Type = "redis"
		applyRedisURL(&cfg)
	}
	if cfg.Type == "" {
		cfg.Type = "memory"
	}
	return cfg
}

// NewStore creates the configured game store.
func NewStore(cfg StoreConfig) store.GameStore {
	if strings.EqualFold(cfg.Type, "redis") || cfg.RedisURL != "" {
		redisStore, err := store.NewRedisStore(
			cfg.RedisAddr,
			cfg.RedisPassword,
			cfg.RedisDB,
			cfg.RedisPoolSize,
			"",
		)
		if err == nil {
			return redisStore
		}
		_, _ = fmt.Fprintf(os.Stderr, "warning: Redis unavailable (%v); falling back to in-memory store\n", err)
	}
	return store.NewMemoryStore()
}

func applyRedisURL(cfg *StoreConfig) {
	parsed, err := url.Parse(cfg.RedisURL)
	if err != nil {
		return
	}
	if parsed.Host != "" {
		cfg.RedisAddr = parsed.Host
	}
	if password, ok := parsed.User.Password(); ok {
		cfg.RedisPassword = password
	}
	if parsed.Path != "" && parsed.Path != "/" {
		db, err := strconv.Atoi(strings.TrimPrefix(parsed.Path, "/"))
		if err == nil {
			cfg.RedisDB = db
		}
	}
}

func envString(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
