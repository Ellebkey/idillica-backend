// user_context.go ≈ user-context.util.ts. The auth middleware stores the JWT
// claims in the gin.Context (≈ req.user); these helpers read them back.
package utils

import (
	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/dto"
)

// UserContextKey is where the auth middleware stores the claims (≈ req.user).
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

// RequireUserID ≈ requireUserId(): 403 when there is no authenticated user.
func RequireUserID(c *gin.Context) (string, error) {
	claims, ok := CurrentUser(c)
	if !ok || claims.ID == "" {
		return "", apperrors.NewForbidden("User authentication required")
	}
	return claims.ID, nil
}
