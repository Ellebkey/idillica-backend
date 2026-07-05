// Package services mirrors src/services.
// jwt.go ≈ jwt.service.ts: access tokens (15 min) + refresh tokens stored in
// Redis under the SAME key scheme as the Node backend (refresh_token:<sha256>,
// refresh_tokens_user:<id>), so sessions are interchangeable between both.
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

// RefreshTokenData is stored as JSON in Redis — same field names as Node.
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

// GenerateToken ≈ generateToken: signs {id, username, roles} with HS256.
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

// GenerateTokenResponse ≈ generateTokenResponse.
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

// CreateRefreshToken ≈ createRefreshToken: random token, hash stored in Redis
// with TTL, plus a per-user set to support "revoke all".
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

// ValidateRefreshToken ≈ validateRefreshToken.
func (s *JWTService) ValidateRefreshToken(ctx context.Context, rawToken string) (*RefreshTokenData, error) {
	stored, err := s.redis.Get(ctx, "refresh_token:"+hashRefreshToken(rawToken)).Result()
	if err != nil {
		// redis.Nil = key not found (≈ a null from redisClient.get)
		return nil, apperrors.NewBusinessRule("Invalid or expired refresh token")
	}

	var data RefreshTokenData
	if err := json.Unmarshal([]byte(stored), &data); err != nil {
		return nil, apperrors.NewBusinessRule("Invalid or expired refresh token")
	}
	return &data, nil
}

// RotateRefreshToken ≈ rotateRefreshToken: burn the old one, issue a new one.
func (s *JWTService) RotateRefreshToken(ctx context.Context, oldRawToken, userID string, rememberMe bool) (string, error) {
	oldHash := hashRefreshToken(oldRawToken)
	s.redis.Del(ctx, "refresh_token:"+oldHash)
	s.redis.SRem(ctx, "refresh_tokens_user:"+userID, oldHash)
	return s.CreateRefreshToken(ctx, userID, rememberMe)
}

// RevokeRefreshToken ≈ revokeRefreshToken. Returns the userID or "" if unknown.
func (s *JWTService) RevokeRefreshToken(ctx context.Context, rawToken string) (string, error) {
	tokenHash := hashRefreshToken(rawToken)
	stored, err := s.redis.Get(ctx, "refresh_token:"+tokenHash).Result()
	if err != nil {
		return "", nil // unknown token: nothing to revoke (mirror: returns null)
	}

	var data RefreshTokenData
	if err := json.Unmarshal([]byte(stored), &data); err != nil {
		return "", nil
	}

	s.redis.Del(ctx, "refresh_token:"+tokenHash)
	s.redis.SRem(ctx, "refresh_tokens_user:"+data.UserID, tokenHash)
	return data.UserID, nil
}

// RevokeAllUserRefreshTokens ≈ revokeAllUserRefreshTokens (pipeline = multi).
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

// ValidatePassword ≈ validatePassword (bcrypt.compare).
func (s *JWTService) ValidatePassword(plaintext, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}

// ValidateToken ≈ validateToken (jwt.verify): parses and checks signature+exp.
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
