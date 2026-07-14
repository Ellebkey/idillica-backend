// logger.go: logging estructurado con log/slog (librería estándar).
// Desarrollo: texto legible a nivel debug. Producción: JSON estructurado.
package config

import (
	"log/slog"
	"os"

	"idilica-backend-go/internal/reqid"
)

// NewLogger construye el logger de la aplicación.
func NewLogger(cfg *Config) *slog.Logger {
	var base slog.Handler

	if cfg.IsProduction() {
		// Producción: JSON, info y superior. El ruido rutinario (p. ej. las
		// sondas 404 del monitor de uptime) se emite a nivel debug y por lo
		// tanto queda descartado aquí.
		base = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		base = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	// reqid.Handler agrega el id de correlación (si el context lo trae) a
	// cada línea; .With fija la etiqueta de la app.
	return slog.New(reqid.NewHandler(base)).With("app", "idilica-api")
}
