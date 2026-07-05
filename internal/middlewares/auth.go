// Package middlewares mirrors src/middlewares.
// auth.go ≈ auth.ts: validates the JWT from the Authorization header and
// stores the claims in the context (≈ req.user = payload).
package middlewares

import (
	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/services"
	"idilica-backend-go/internal/utils"
)

type Auth struct {
	jwtService *services.JWTService
}

func NewAuth(jwtService *services.JWTService) *Auth {
	return &Auth{jwtService: jwtService}
}

// CheckAuth ≈ Auth.checkAuth. The backend expects the RAW token in the
// Authorization header — no "Bearer" prefix — exactly like the Node app
// (the React frontend already sends it that way).
func (a *Auth) CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token == "" {
			_ = c.Error(apperrors.NewUnauthorized("Authentication token is required"))
			c.Abort()
			return
		}

		claims, err := a.jwtService.ValidateToken(token)
		if err != nil {
			// 401 so the frontend interceptor triggers a token refresh
			_ = c.Error(apperrors.NewUnauthorized("Token authentication failed"))
			c.Abort()
			return
		}

		c.Set(utils.UserContextKey, claims)
		c.Next()
	}
}
