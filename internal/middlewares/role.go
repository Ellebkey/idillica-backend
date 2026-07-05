// role.go: guarda de roles a nivel aplicación (p. ej. 'admin').
// Los permisos por cocina (owner/editor/viewer) viven en los services.
package middlewares

import (
	"slices"

	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/utils"
)

// RequireRole exige alguno de los roles dados. Debe correr DESPUÉS de CheckAuth.
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
