// logger.go ≈ the HTTP request logging block in express.ts (dev/test only):
// logs the incoming request and, on finish, status + duration.
package middlewares

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		logger.Debug("request", "method", c.Request.Method, "url", c.Request.URL.String())

		c.Next()

		logger.Debug("response",
			"method", c.Request.Method,
			"status", c.Writer.Status(),
			"url", c.Request.URL.String(),
			"duration", time.Since(start).String(),
		)
	}
}
