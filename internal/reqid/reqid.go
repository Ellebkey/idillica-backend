// Package reqid propaga un identificador de correlación por request:
// lo toma del header entrante (si es válido) o genera uno, lo guarda en el
// context, y un slog.Handler lo agrega a cada línea de log — así se pueden
// cruzar los logs de una misma petición entre cliente y servidor.
package reqid

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"regexp"
)

// HeaderName es el header estándar de correlación (entrada y salida).
const HeaderName = "X-Request-Id"

// LogKey es la clave con la que el id aparece en cada línea de log.
const LogKey = "request_id"

// valid acota los ids entrantes que aceptamos tal cual: cortos y sin
// caracteres raros. Cualquier otra cosa se reemplaza por un UUID generado.
var valid = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)

type ctxKey struct{}

// Sanitize devuelve el id entrante si es válido, o uno nuevo si no lo es.
func Sanitize(incoming string) string {
	if valid.MatchString(incoming) {
		return incoming
	}
	return New()
}

// WithID guarda el id de correlación en el context.
func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext recupera el id de correlación (cadena vacía si no hay).
func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// New genera un UUID v4 con crypto/rand.
func New() string {
	var b [16]byte
	_, _ = rand.Read(b[:])      // crypto/rand.Read no falla en Go 1.24+
	b[6] = (b[6] & 0x0f) | 0x40 // versión 4
	b[8] = (b[8] & 0x3f) | 0x80 // variante RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Handler envuelve otro slog.Handler y, si el context trae un id de
// correlación, lo agrega como atributo a cada registro.
type Handler struct {
	slog.Handler
}

// NewHandler envuelve h para inyectar el id de correlación del context.
func NewHandler(h slog.Handler) *Handler {
	return &Handler{Handler: h}
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	if id := FromContext(ctx); id != "" {
		r.AddAttrs(slog.String(LogKey, id))
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs y WithGroup re-envuelven el resultado para no perder la
// inyección del id al derivar loggers (p. ej. logger.With(...)).
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{Handler: h.Handler.WithGroup(name)}
}
