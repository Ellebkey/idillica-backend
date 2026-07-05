// logger.go ≈ logger.ts: winston → log/slog (standard library, Go 1.21+).
// Development: human-readable text at debug level. Production: structured JSON.
package config

import (
	"log/slog"
	"os"
)

// NewLogger builds the application logger, mirroring the dev/prod split of
// the winston setup (devFormat printf vs prodFormat json).
func NewLogger(cfg *Config) *slog.Logger {
	var handler slog.Handler

	if cfg.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	// `label: 'idilica-api'` of the winston config
	return slog.New(handler).With("app", "idilica-api")
}
