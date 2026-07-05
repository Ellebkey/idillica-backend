// Package controllers: handlers delgados — bindean el DTO, llaman al service
// y escriben la respuesta; los errores viajan con c.Error(err).
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/services"
	"idilica-backend-go/internal/utils"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (ac *AuthController) Login(c *gin.Context) {
	d, err := utils.BindJSON[dto.LoginDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response, err := ac.authService.Login(c.Request.Context(), d)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (ac *AuthController) Register(c *gin.Context) {
	d, err := utils.BindJSON[dto.RegisterDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := ac.authService.Register(c.Request.Context(), d); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Cuenta creada correctamente.",
	})
}

func (ac *AuthController) ChangePassword(c *gin.Context) {
	d, err := utils.BindJSON[dto.ChangePasswordDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	claims, _ := utils.CurrentUser(c)
	// en esta app el username ES el correo
	if err := ac.authService.ChangePassword(c.Request.Context(), claims.Username, d); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password changed successfully"})
}

func (ac *AuthController) ResetPassword(c *gin.Context) {
	d, err := utils.BindJSON[dto.ResetPasswordDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := ac.authService.ResetPassword(c.Request.Context(), d.Email); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "If the email exists, a reset link has been sent",
	})
}

func (ac *AuthController) ConfirmResetPassword(c *gin.Context) {
	d, err := utils.BindJSON[dto.ConfirmResetPasswordDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := ac.authService.ConfirmResetPassword(c.Request.Context(), d); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password has been reset successfully"})
}

func (ac *AuthController) VerifyEmail(c *gin.Context) {
	d, err := utils.BindJSON[dto.VerifyEmailDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := ac.authService.VerifyEmail(c.Request.Context(), d); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Email verified successfully"})
}

func (ac *AuthController) ResendVerificationEmail(c *gin.Context) {
	d, err := utils.BindJSON[dto.ResetPasswordDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := ac.authService.ResendVerificationEmail(c.Request.Context(), d.Email); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "If the email exists, a verification link has been sent",
	})
}

func (ac *AuthController) Refresh(c *gin.Context) {
	d, err := utils.BindJSON[dto.RefreshTokenDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response, err := ac.authService.Refresh(c.Request.Context(), d)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (ac *AuthController) Logout(c *gin.Context) {
	d, err := utils.BindJSON[dto.LogoutDto](c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := ac.authService.Logout(c.Request.Context(), d); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logged out successfully"})
}
