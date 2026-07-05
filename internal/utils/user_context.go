// user_context.go: el middleware de auth guarda los claims del JWT en el
// gin.Context; estos helpers los leen de vuelta.
package utils

import (
	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/dto"
)

// UserContextKey es la llave donde el middleware de auth guarda los claims.
const UserContextKey = "user"

// CurrentUser returns the authenticated user's claims, if any.
func CurrentUser(c *gin.Context) (*dto.JWTClaims, bool) {
	value, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}
	claims, ok := value.(*dto.JWTClaims)
	return claims, ok
}

// RequireUserID: 403 cuando no hay usuario autenticado.
func RequireUserID(c *gin.Context) (string, error) {
	claims, ok := CurrentUser(c)
	if !ok || claims.ID == "" {
		return "", apperrors.NewForbidden("User authentication required")
	}
	return claims.ID, nil
}
