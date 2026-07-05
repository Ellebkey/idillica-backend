// logger.go: log de requests HTTP (solo dev/test) — la entrada y, al
// terminar, status + duración.
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
