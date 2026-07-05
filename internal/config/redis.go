// redis.go ≈ redis-config.ts. Same role: refresh tokens and one-time tokens
// (email verify / password reset). Key scheme is identical to the Node app,
// so sessions are interchangeable between both backends.
package config

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedis creates the client and pings it. Like the Node app, a Redis that is
// down does not prevent boot — flows that need it will fail with clear errors.
func NewRedis(cfg *Config, logger *slog.Logger) *redis.Client {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Error("Redis client Error", "error", err)
		// Fall back to defaults so the app can still boot
		opts = &redis.Options{Addr: "localhost:6379"}
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("Redis client Error", "error", err)
	} else {
		logger.Info("Redis connected")
	}

	return client
}
