// redis.go: cliente Redis para refresh tokens y tokens de un solo uso
// (verificación de correo / reset de contraseña).
package config

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedis crea el cliente y hace ping. Un Redis caído no impide el arranque:
// los flujos que lo necesitan fallarán con errores claros.
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
