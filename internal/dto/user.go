// user.go ≈ user.dto.ts — auth/user DTOs and JWT payloads.
// json tags keep the exact field names of the Node API (camelCase).
package dto

import "github.com/golang-jwt/jwt/v5"

// ===== AUTH DTOs (binding ≈ auth.validation.ts) =====

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

// JWTPayload ≈ the payload signed by the Node backend ({id, username, roles}).
// Same JWT_SECRET ⇒ tokens are interchangeable between both backends.
type JWTPayload struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

// JWTClaims is the payload plus the registered claims (exp). golang-jwt
// validates `exp` automatically on parse, like jsonwebtoken.verify.
type JWTClaims struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTResponse ≈ JWTResponse of user.dto.ts — the login/refresh body.
type JWTResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	Roles        []string `json:"roles"`
	Username     string   `json:"username"`
	Fullname     *string  `json:"fullname,omitempty"`
	ExpiresIn    string   `json:"expiresIn"` // ISO timestamp, like formatISO(addMinutes(now, 15))
}
