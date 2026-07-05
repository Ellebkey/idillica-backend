// Package services contiene la lógica de negocio.
// jwt.go: access tokens (15 min) + refresh tokens en Redis
// (refresh_token:<sha256>, refresh_tokens_user:<id>).
package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"idilica-backend-go/internal/apperrors"
	"idilica-backend-go/internal/dto"
)

const (
	accessTokenMinutes = 15
	refreshTtlShort    = 8 * time.Hour       // rememberMe = false
	refreshTtlLong     = 30 * 24 * time.Hour // rememberMe = true
)

// RefreshTokenData se guarda como JSON en Redis.
type RefreshTokenData struct {
	UserID     string `json:"userId"`
	RememberMe bool   `json:"rememberMe"`
	CreatedAt  string `json:"createdAt"`
}

type JWTService struct {
	secret string
	redis  *redis.Client
	logger *slog.Logger
}

func NewJWTService(secret string, redisClient *redis.Client, logger *slog.Logger) *JWTService {
	return &JWTService{secret: secret, redis: redisClient, logger: logger}
}

// GenerateToken firma {id, username, roles} con HS256.
func (s *JWTService) GenerateToken(payload dto.JWTPayload) (string, error) {
	claims := dto.JWTClaims{
		ID:       payload.ID,
		Username: payload.Username,
		Roles:    payload.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenMinutes * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.secret))
	if err != nil {
		s.logger.Error("Error generating JWT token", "error", err)
		return "", apperrors.NewBusinessRule("Failed to generate authentication token")
	}

	s.logger.Info("JWT token generated", "username", payload.Username)
	return token, nil
}

// GenerateTokenResponse arma el cuerpo completo de login/refresh.
func (s *JWTService) GenerateTokenResponse(ctx context.Context, user dto.JWTPayload, rememberMe bool) (*dto.JWTResponse, error) {
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.CreateRefreshToken(ctx, user.ID, rememberMe)
	if err != nil {
		return nil, err
	}

	return &dto.JWTResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Roles:        user.Roles,
		Username:     user.Username,
		ExpiresIn:    time.Now().Add(accessTokenMinutes * time.Minute).Format(time.RFC3339),
	}, nil
}

// CreateRefreshToken: token aleatorio cuyo hash se guarda en Redis con TTL,
// más un set por usuario para poder "revocar todos".
func (s *JWTService) CreateRefreshToken(ctx context.Context, userID string, rememberMe bool) (string, error) {
	raw := make([]byte, 48)
	if _, err := rand.Read(raw); err != nil {
		return "", apperrors.NewBusinessRule("Failed to generate refresh token")
	}
	rawToken := hex.EncodeToString(raw)
	tokenHash := hashRefreshToken(rawToken)

	ttl := refreshTtlShort
	if rememberMe {
		ttl = refreshTtlLong
	}

	data, err := json.Marshal(RefreshTokenData{
		UserID:     userID,
		RememberMe: rememberMe,
		CreatedAt:  time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}

	if err := s.redis.Set(ctx, "refresh_token:"+tokenHash, data, ttl).Err(); err != nil {
		return "", err
	}
	if err := s.redis.SAdd(ctx, "refresh_tokens_user:"+userID, tokenHash).Err(); err != nil {
		return "", err
	}
	if err := s.redis.Expire(ctx, "refresh_tokens_user:"+userID, refreshTtlLong).Err(); err != nil {
		return "", err
	}

	return rawToken, nil
}

// ValidateRefreshToken verifica y devuelve los datos del refresh token.
func (s *JWTService) ValidateRefreshToken(ctx context.Context, rawToken string) (*RefreshTokenData, error) {
	stored, err := s.redis.Get(ctx, "refresh_token:"+hashRefreshToken(rawToken)).Result()
	if err != nil {
		// redis.Nil = llave inexistente
		return nil, apperrors.NewBusinessRule("Invalid or expired refresh token")
	}

	var data RefreshTokenData
	if err := json.Unmarshal([]byte(stored), &data); err != nil {
		return nil, apperrors.NewBusinessRule("Invalid or expired refresh token")
	}
	return &data, nil
}

// RotateRefreshToken quema el token anterior y emite uno nuevo.
func (s *JWTService) RotateRefreshToken(ctx context.Context, oldRawToken, userID string, rememberMe bool) (string, error) {
	oldHash := hashRefreshToken(oldRawToken)
	s.redis.Del(ctx, "refresh_token:"+oldHash)
	s.redis.SRem(ctx, "refresh_tokens_user:"+userID, oldHash)
	return s.CreateRefreshToken(ctx, userID, rememberMe)
}

// RevokeRefreshToken revoca; devuelve el userID o "" si no existía.
func (s *JWTService) RevokeRefreshToken(ctx context.Context, rawToken string) (string, error) {
	tokenHash := hashRefreshToken(rawToken)
	stored, err := s.redis.Get(ctx, "refresh_token:"+tokenHash).Result()
	if err != nil {
		return "", nil // token desconocido: nada que revocar
	}

	var data RefreshTokenData
	if err := json.Unmarshal([]byte(stored), &data); err != nil {
		return "", nil
	}

	s.redis.Del(ctx, "refresh_token:"+tokenHash)
	s.redis.SRem(ctx, "refresh_tokens_user:"+data.UserID, tokenHash)
	return data.UserID, nil
}

// RevokeAllUserRefreshTokens borra todas las sesiones del usuario.
func (s *JWTService) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	setKey := "refresh_tokens_user:" + userID
	tokenHashes, err := s.redis.SMembers(ctx, setKey).Result()
	if err != nil {
		return err
	}

	if len(tokenHashes) > 0 {
		pipe := s.redis.TxPipeline()
		for _, hash := range tokenHashes {
			pipe.Del(ctx, "refresh_token:"+hash)
		}
		pipe.Del(ctx, setKey)
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
	}

	s.logger.Info("All refresh tokens revoked", "userId", userID, "count", len(tokenHashes))
	return nil
}

// ValidatePassword compara la contraseña contra su hash bcrypt.
func (s *JWTService) ValidatePassword(plaintext, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}

// ValidateToken parsea el token y verifica firma + expiración.
func (s *JWTService) ValidateToken(tokenString string) (*dto.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &dto.JWTClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(s.secret), nil
	})
	if err != nil || !token.Valid {
		return nil, apperrors.NewBusinessRule("Invalid or expired token")
	}

	claims, ok := token.Claims.(*dto.JWTClaims)
	if !ok {
		return nil, apperrors.NewBusinessRule("Invalid or expired token")
	}
	return claims, nil
}

func hashRefreshToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
