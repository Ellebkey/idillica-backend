// role.go ≈ role.middleware.ts: guards app-level roles (e.g. 'admin').
// Per-cocina permissions (owner/editor/viewer) live in the services.
package middlewares

import (
	"slices"

	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/utils"
)

// RequireRole ≈ requireRole(...roles). Must run AFTER CheckAuth.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := utils.CurrentUser(c)
		if !ok || !hasAny(claims.Roles, roles) {
			_ = c.Error(apperrors.NewForbidden("Access denied: insufficient privileges"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func hasAny(userRoles, wanted []string) bool {
	for _, role := range wanted {
		if slices.Contains(userRoles, role) {
			return true
		}
	}
	return false
}
