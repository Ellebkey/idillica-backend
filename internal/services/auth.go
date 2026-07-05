// auth.go: registro, login y ciclo de vida de contraseñas y sesiones.
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/config"
	"idilica-backend-go/internal/dto"
	"idilica-backend-go/internal/models"
)

const (
	resetTokenTtl  = time.Hour
	verifyTokenTtl = 24 * time.Hour
	saltRounds     = 10 // costo bcrypt
)

type AuthService struct {
	db     *gorm.DB
	redis  *redis.Client
	cfg    *config.Config
	jwt    *JWTService
	email  *EmailService
	logger *slog.Logger
}

func NewAuthService(
	db *gorm.DB,
	redisClient *redis.Client,
	cfg *config.Config,
	jwtService *JWTService,
	emailService *EmailService,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{db: db, redis: redisClient, cfg: cfg, jwt: jwtService, email: emailService, logger: logger}
}

// Login valida credenciales y emite tokens.
func (s *AuthService) Login(ctx context.Context, d *dto.LoginDto) (*dto.JWTResponse, error) {
	invalidCredentials := apperrors.NewBusinessRule("Invalid username or password")

	var user models.User
	err := s.db.WithContext(ctx).First(&user, "username = ?", d.Username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, invalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if !s.jwt.ValidatePassword(d.Password, user.HashedPassword) {
		return nil, invalidCredentials
	}

	if s.cfg.RequireEmailVerification && !user.EmailVerified {
		return nil, apperrors.NewBusinessRule("Please verify your email before logging in")
	}

	roles := []string(user.Roles)
	if roles == nil {
		roles = []string{}
	}

	response, err := s.jwt.GenerateTokenResponse(ctx, dto.JWTPayload{
		ID:       user.ID,
		Username: user.Username,
		Roles:    roles,
	}, d.RememberMe)
	if err != nil {
		return nil, err
	}

	response.Fullname = user.Fullname
	return response, nil
}

// Refresh valida y rota el refresh token.
func (s *AuthService) Refresh(ctx context.Context, d *dto.RefreshTokenDto) (*dto.JWTResponse, error) {
	tokenData, err := s.jwt.ValidateRefreshToken(ctx, d.RefreshToken)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", tokenData.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFound("User", tokenData.UserID)
		}
		return nil, err
	}

	newRefreshToken, err := s.jwt.RotateRefreshToken(ctx, d.RefreshToken, user.ID, tokenData.RememberMe)
	if err != nil {
		return nil, err
	}

	token, err := s.jwt.GenerateToken(dto.JWTPayload{ID: user.ID, Username: user.Username, Roles: user.Roles})
	if err != nil {
		return nil, err
	}

	return &dto.JWTResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		Roles:        user.Roles,
		Username:     user.Username,
		ExpiresIn:    time.Now().Add(accessTokenMinutes * time.Minute).Format(time.RFC3339),
	}, nil
}

// Logout revoca el refresh token.
func (s *AuthService) Logout(ctx context.Context, d *dto.LogoutDto) error {
	userID, err := s.jwt.RevokeRefreshToken(ctx, d.RefreshToken)
	if err != nil {
		return err
	}
	if userID != "" {
		s.logger.Info("User logged out successfully", "userId", userID)
	}
	return nil
}

// Register crea el usuario Y su cocina personal (como owner) en una sola
// transacción. Con la verificación apagada, la cuenta queda lista para entrar.
func (s *AuthService) Register(ctx context.Context, d *dto.RegisterDto) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		hashedPassword, err := HashPassword(d.Password)
		if err != nil {
			return err
		}

		roles := d.Roles
		if roles == nil {
			roles = []string{}
		}

		var fullname *string
		if trimmed := strings.TrimSpace(d.Fullname); trimmed != "" {
			fullname = &trimmed
		}

		user := models.User{
			Username:       d.Username,
			Email:          d.Email,
			Fullname:       fullname,
			HashedPassword: hashedPassword,
			EmailVerified:  !s.cfg.RequireEmailVerification,
			Roles:          models.Roles(roles),
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		cocinaName := "Mi cocina"
		if fullname != nil {
			cocinaName = *fullname
		}
		cocina := models.Cocina{
			Name:             cocinaName,
			Moneda:           "MXN",
			ImpuestoDefault:  0.16,
			FoodCostObjetivo: 0.30,
		}
		if err := tx.Create(&cocina).Error; err != nil {
			return err
		}

		member := models.CocinaMember{CocinaID: cocina.ID, UserID: user.ID, Rol: models.RolOwner}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		s.logger.Info("User registered", "userId", user.ID, "username", user.Username)

		if s.cfg.RequireEmailVerification {
			verifyToken, err := randomToken()
			if err != nil {
				return err
			}
			if err := s.redis.Set(ctx, "email_verify:"+verifyToken, user.ID, verifyTokenTtl).Err(); err != nil {
				return err
			}
			if err := s.email.SendEmailVerification(d.Email, verifyToken); err != nil {
				return err
			}
			s.logger.Info("Verification email sent", "email", d.Email)
		}

		return nil
	})
}

// ChangePassword — el email sale del username del JWT (en esta app el
// username ES el correo).
func (s *AuthService) ChangePassword(ctx context.Context, email string, d *dto.ChangePasswordDto) error {
	var userID string

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, "email = ?", email).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.NewNotFound("User", email)
			}
			return err
		}

		if !s.jwt.ValidatePassword(d.CurrentPassword, user.HashedPassword) {
			return apperrors.NewBusinessRule("Current password is incorrect")
		}

		hashedPassword, err := HashPassword(d.NewPassword)
		if err != nil {
			return err
		}
		if err := tx.Model(&user).Update("hashed_password", hashedPassword).Error; err != nil {
			return err
		}
		userID = user.ID
		return nil
	})
	if err != nil {
		return err
	}

	if userID != "" {
		if err := s.jwt.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
			return err
		}
	}

	s.logger.Info("Password changed successfully", "email", email)
	return nil
}

// ResetPassword: silencioso si el correo no existe (evita enumeración).
func (s *AuthService) ResetPassword(ctx context.Context, email string) error {
	var user models.User
	err := s.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Warn("Password reset requested for non-existent email", "email", email)
		return nil
	}
	if err != nil {
		return err
	}

	resetToken, err := randomToken()
	if err != nil {
		return err
	}
	if err := s.redis.Set(ctx, "password_reset:"+resetToken, user.ID, resetTokenTtl).Err(); err != nil {
		return err
	}
	if err := s.email.SendPasswordResetEmail(email, resetToken); err != nil {
		return err
	}

	s.logger.Info("Password reset token generated", "email", email)
	return nil
}

// ConfirmResetPassword aplica la nueva contraseña con el token de un solo uso.
func (s *AuthService) ConfirmResetPassword(ctx context.Context, d *dto.ConfirmResetPasswordDto) error {
	redisKey := "password_reset:" + d.Token
	userID, err := s.redis.Get(ctx, redisKey).Result()
	if err != nil {
		return apperrors.NewBusinessRule("Invalid or expired reset token")
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.NewNotFound("User", userID)
			}
			return err
		}

		hashedPassword, err := HashPassword(d.NewPassword)
		if err != nil {
			return err
		}
		return tx.Model(&user).Update("hashed_password", hashedPassword).Error
	})
	if err != nil {
		return err
	}

	s.redis.Del(ctx, redisKey)
	if err := s.jwt.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
		return err
	}

	s.logger.Info("Password reset confirmed", "userId", userID)
	return nil
}

// VerifyEmail marca el correo como verificado.
func (s *AuthService) VerifyEmail(ctx context.Context, d *dto.VerifyEmailDto) error {
	redisKey := "email_verify:" + d.Token
	userID, err := s.redis.Get(ctx, redisKey).Result()
	if err != nil {
		return apperrors.NewBusinessRule("Invalid or expired verification token")
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.NewNotFound("User", userID)
			}
			return err
		}
		if user.EmailVerified {
			return apperrors.NewBusinessRule("Email is already verified")
		}
		return tx.Model(&user).Update("email_verified", true).Error
	})
	if err != nil {
		return err
	}

	s.redis.Del(ctx, redisKey)
	s.logger.Info("Email verified", "userId", userID)
	return nil
}

// ResendVerificationEmail reenvía el token de verificación.
func (s *AuthService) ResendVerificationEmail(ctx context.Context, email string) error {
	var user models.User
	err := s.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil // silencioso: evita enumeración de correos
	}
	if err != nil {
		return err
	}

	if user.EmailVerified {
		return apperrors.NewBusinessRule("Email is already verified")
	}

	verifyToken, err := randomToken()
	if err != nil {
		return err
	}
	if err := s.redis.Set(ctx, "email_verify:"+verifyToken, user.ID, verifyTokenTtl).Err(); err != nil {
		return err
	}
	if err := s.email.SendEmailVerification(email, verifyToken); err != nil {
		return err
	}

	s.logger.Info("Verification email resent", "email", email)
	return nil
}

// HashPassword — función suelta para poder probarla sin armar el service.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), saltRounds)
	if err != nil {
		return "", apperrors.NewBusinessRule("Failed to hash password")
	}
	return string(hashed), nil
}

// randomToken genera 32 bytes aleatorios codificados en hex.
func randomToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
