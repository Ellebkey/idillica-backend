// middleware.go ≈ errors/middleware.ts: normalizes any error into an AppError
// and writes the exact same JSON shape as the Node backend:
//
//	{ "error": { "code", "message", "status", "details"?, "stack"? } }
//
// Pattern: controllers call c.Error(err) (≈ next(error)) and this middleware,
// registered first, inspects c.Errors after c.Next().
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

// Normalize ≈ normalizeError(): AppError passes through; database errors are
// translated (unique violation → 409, like SequelizeUniqueConstraintError);
// anything else becomes a 500.
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
	// Stack trace only in development (mirror of `envConfig.env === 'development'`)
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

// ErrorHandler ≈ converterErr + errorMiddleware.
func ErrorHandler(env string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}
		respond(c, env, logger, Normalize(c.Errors.Last().Err))
	}
}

// NotFoundHandler ≈ notFound(): unmatched routes, same shape and message.
func NotFoundHandler(env string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		respond(c, env, logger, &AppError{
			Code:       "ROUTE_NOT_FOUND",
			Message:    fmt.Sprintf("Cannot %s %s", c.Request.Method, c.Request.URL.Path),
			StatusCode: http.StatusNotFound,
		})
	}
}

// Recovery turns panics into the standard 500 JSON (Express does this via the
// error middleware; in Gin a panic needs an explicit recovery handler).
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
