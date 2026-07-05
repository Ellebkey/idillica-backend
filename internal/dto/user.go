// user.go — DTOs de auth/usuario y payloads JWT (json en camelCase).
package dto

import "github.com/golang-jwt/jwt/v5"

// ===== AUTH DTOs =====

type LoginDto struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"rememberMe"`
}

type RegisterDto struct {
	Username string   `json:"username" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8,max=128"`
	Email    string   `json:"email" binding:"required,email"`
	Fullname string   `json:"fullname" binding:"omitempty,max=100"`
	Roles    []string `json:"roles"`
}

type ChangePasswordDto struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=128"`
}

type ResetPasswordDto struct {
	Email string `json:"email" binding:"required,email"`
}

type ConfirmResetPasswordDto struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8,max=128"`
}

type VerifyEmailDto struct {
	Token string `json:"token" binding:"required"`
}

type RefreshTokenDto struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutDto struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// ===== JWT DTOs =====

// JWTPayload — lo que se firma dentro del access token.
type JWTPayload struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

// JWTClaims — el payload más los registered claims (exp); golang-jwt valida
// la expiración automáticamente al parsear.
type JWTClaims struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTResponse — cuerpo de respuesta de login/refresh.
type JWTResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	Roles        []string `json:"roles"`
	Username     string   `json:"username"`
	Fullname     *string  `json:"fullname,omitempty"`
	ExpiresIn    string   `json:"expiresIn"` // timestamp ISO de expiración del access token
}
