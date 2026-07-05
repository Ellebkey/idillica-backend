// logger.go: logging estructurado con log/slog (librería estándar).
// Desarrollo: texto legible a nivel debug. Producción: JSON estructurado.
package config

import (
	"log/slog"
	"os"
)

// NewLogger construye el logger de la aplicación.
func NewLogger(cfg *Config) *slog.Logger {
	var handler slog.Handler

	if cfg.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	// etiqueta fija de la app en cada línea de log
	return slog.New(handler).With("app", "idilica-api")
}
