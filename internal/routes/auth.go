// auth.go ≈ auth.route.ts — same paths, same rate limiting.
package routes

import (
	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/config"
	"idilica-backend-go/internal/controllers"
	"idilica-backend-go/internal/middlewares"
)

func registerAuthRoutes(
	api *gin.RouterGroup,
	ctrl *controllers.AuthController,
	auth *middlewares.Auth,
	cfg *config.Config,
) {
	limiter := middlewares.AuthRateLimiter(cfg.Env)

	api.POST("/auth/login", limiter, ctrl.Login)
	api.POST("/auth/register", limiter, ctrl.Register)
	api.POST("/auth/change-password", auth.CheckAuth(), ctrl.ChangePassword)
	api.POST("/auth/reset-password", limiter, ctrl.ResetPassword)
	api.POST("/auth/confirm-reset-password", limiter, ctrl.ConfirmResetPassword)
	api.POST("/auth/verify-email", limiter, ctrl.VerifyEmail)
	api.POST("/auth/resend-verification", limiter, ctrl.ResendVerificationEmail)
	api.POST("/auth/refresh", limiter, ctrl.Refresh)
	api.POST("/auth/logout", limiter, ctrl.Logout)
}
