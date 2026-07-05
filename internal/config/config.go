// Package config carga y valida la configuración del servicio.
// config.go: lee .env, valida las variables requeridas y expone una config tipada.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config es la configuración tipada de todo el servicio.
type Config struct {
	Env                      string // APP_ENV: development | production | stage | test
	Port                     int
	JWTSecret                string
	FrontendURL              string
	RequireEmailVerification bool
	ResendAPIKey             string
	ResendFromEmail          string
	SQLHost                  string
	SQLDB                    string
	SQLUser                  string
	SQLPassword              string
	SQLPort                  int
	RedisURL                 string
	MaxPool                  int
	MinPool                  int
}

// Load lee .env (opcional) y valida el entorno: si falta una variable
// requerida, el arranque falla de inmediato con la lista completa.
func Load() (*Config, error) {
	// .env del working directory; se ignora en silencio si no existe
	_ = godotenv.Load()

	cfg := &Config{
		Env:                      getEnv("APP_ENV", "development"),
		Port:                     getEnvInt("PORT", 4051),
		JWTSecret:                os.Getenv("JWT_SECRET"),
		FrontendURL:              getEnv("FRONTEND_URL", "http://localhost:5273"),
		RequireEmailVerification: getEnvBool("AUTH_REQUIRE_EMAIL_VERIFICATION", false),
		ResendAPIKey:             os.Getenv("RESEND_API_KEY"),
		ResendFromEmail:          getEnv("RESEND_FROM_EMAIL", "noreply@idilica.app"),
		SQLHost:                  os.Getenv("SQL_HOST"),
		SQLDB:                    os.Getenv("SQL_DB"),
		SQLUser:                  os.Getenv("SQL_USER"),
		SQLPassword:              os.Getenv("SQL_PASSWORD"),
		SQLPort:                  getEnvInt("SQL_PORT", 5432),
		RedisURL:                 getEnv("REDIS_URL", "redis://localhost:6379"),
		MaxPool:                  10,
		MinPool:                  1,
	}

	validEnvs := map[string]bool{"development": true, "production": true, "stage": true, "test": true}
	if !validEnvs[cfg.Env] {
		return nil, fmt.Errorf("config validation error: APP_ENV must be one of development|production|stage|test, got %q", cfg.Env)
	}

	// Variables obligatorias
	missing := []string{}
	required := map[string]string{
		"JWT_SECRET":   cfg.JWTSecret,
		"SQL_HOST":     cfg.SQLHost,
		"SQL_DB":       cfg.SQLDB,
		"SQL_USER":     cfg.SQLUser,
		"SQL_PASSWORD": cfg.SQLPassword,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("config validation error: missing required env vars: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

// IsProduction agrupa production y stage.
func (c *Config) IsProduction() bool {
	return c.Env == "production" || c.Env == "stage"
}

func (c *Config) IsDevelopment() bool { return c.Env == "development" }
func (c *Config) IsTest() bool        { return c.Env == "test" }

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return fallback
}
