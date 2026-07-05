// Package middlewares: piezas transversales de la cadena HTTP.
// auth.go valida el JWT del header Authorization y guarda los claims en el
// contexto del request.
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

// CheckAuth espera el token CRUDO en el header Authorization — sin prefijo
// "Bearer" — que es como lo envía el frontend.
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
