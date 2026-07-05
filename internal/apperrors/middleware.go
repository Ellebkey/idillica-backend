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

	logger.Error(appErr.Message,
		"code", appErr.Code,
		"status", appErr.StatusCode,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)

	c.JSON(appErr.StatusCode, gin.H{"error": body})
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
