// middleware.go: normaliza cualquier error a AppError y escribe la respuesta
// JSON estándar del API:
//
//	{ "error": { "code", "message", "status", "details"?, "stack"? } }
//
// Patrón: los controllers llaman c.Error(err) y este middleware, registrado
// primero, inspecciona c.Errors después de c.Next().
package apperrors

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Normalize: un AppError pasa tal cual; los errores de base de datos se
// traducen (violación de unique → 409); todo lo demás se vuelve 500.
func Normalize(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			msg := "A record with these values already exists"
			if pgErr.Detail != "" {
				msg = fmt.Sprintf("A record already exists: %s", pgErr.Detail)
			}
			return NewConflict(msg)
		case "23503": // foreign_key_violation
			return NewConflict(fmt.Sprintf("Cannot perform operation due to %s constraint", pgErr.ConstraintName))
		}
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &AppError{Code: "NOT_FOUND", Message: "Resource not found", StatusCode: 404}
	}

	internal := NewInternal(err.Error())
	internal.Err = err
	return internal
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Details any    `json:"details,omitempty"`
	Stack   string `json:"stack,omitempty"`
}

func respond(c *gin.Context, env string, logger *slog.Logger, appErr *AppError) {
	body := errorBody{Code: appErr.Code, Message: appErr.Message, Status: appErr.StatusCode}
	if appErr.Details != nil {
		body.Details = appErr.Details
	}
	// Stack trace solo en desarrollo
	if env == "development" && appErr.Err != nil {
		body.Stack = appErr.Err.Error()
	}

	logResult(c, logger, appErr)

	c.JSON(appErr.StatusCode, gin.H{"error": body})
}

// logResult registra el error al nivel adecuado para que los errores de
// cliente dejen de contaminar el stream de errores:
//   - 5xx (fallo real del servidor)      → error
//   - ruta inexistente (sonda de uptime) → debug (se descarta en producción)
//   - resto de 4xx (cliente)             → warn
//
// Usa LogAttrs con el context del request para que el Handler de reqid
// agregue el id de correlación a la línea.
func logResult(c *gin.Context, logger *slog.Logger, appErr *AppError) {
	level := levelFor(appErr)

	attrs := []slog.Attr{
		slog.String("code", appErr.Code),
		slog.Int("status", appErr.StatusCode),
		slog.String("path", c.Request.URL.Path),
		slog.String("method", c.Request.Method),
	}
	// El detalle de la causa solo interesa cuando es un fallo real del servidor.
	if level >= slog.LevelError && appErr.Err != nil {
		attrs = append(attrs, slog.Any("err", appErr.Err))
	}

	logger.LogAttrs(c.Request.Context(), level, appErr.Message, attrs...)
}

func levelFor(appErr *AppError) slog.Level {
	switch {
	case appErr.StatusCode >= 500:
		return slog.LevelError
	case appErr.Code == "ROUTE_NOT_FOUND":
		return slog.LevelDebug
	default:
		return slog.LevelWarn
	}
}

// ErrorHandler escribe la respuesta de error al final de la cadena.
func ErrorHandler(env string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}
		respond(c, env, logger, Normalize(c.Errors.Last().Err))
	}
}

// NotFoundHandler responde rutas inexistentes con la misma forma JSON.
func NotFoundHandler(env string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		respond(c, env, logger, &AppError{
			Code:       "ROUTE_NOT_FOUND",
			Message:    fmt.Sprintf("Cannot %s %s", c.Request.Method, c.Request.URL.Path),
			StatusCode: http.StatusNotFound,
		})
	}
}

// Recovery convierte panics en el 500 JSON estándar.
func Recovery(env string, logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		respond(c, env, logger, &AppError{
			Code:       "INTERNAL_ERROR",
			Message:    "An internal server error occurred",
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("panic: %v", recovered),
		})
		c.Abort()
	})
}
