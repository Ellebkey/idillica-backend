// jwt_test.go — pruebas unitarias del servicio JWT (go test ./...).
package services

import (
	"io"
	"log/slog"
	"testing"

	"idilica-backend-go/internal/dto"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestJWTRoundTrip(t *testing.T) {
	svc := NewJWTService("test-secret", nil, discardLogger())

	token, err := svc.GenerateToken(dto.JWTPayload{
		ID:       "user-1",
		Username: "ana@idilica.app",
		Roles:    []string{"admin"},
	})
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.ID != "user-1" || claims.Username != "ana@idilica.app" {
		t.Errorf("claims mismatch: %+v", claims)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
		t.Errorf("roles mismatch: %v", claims.Roles)
	}
}

func TestValidateTokenRejectsWrongSecret(t *testing.T) {
	issuer := NewJWTService("secret-a", nil, discardLogger())
	verifier := NewJWTService("secret-b", nil, discardLogger())

	token, err := issuer.GenerateToken(dto.JWTPayload{ID: "u", Username: "x@y.z", Roles: nil})
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	if _, err := verifier.ValidateToken(token); err == nil {
		t.Error("expected validation to fail with a different secret")
	}
}

func TestPasswordHashRoundTrip(t *testing.T) {
	svc := NewJWTService("irrelevant", nil, discardLogger())

	hash, err := HashPassword("super-secreto-123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !svc.ValidatePassword("super-secreto-123", hash) {
		t.Error("expected the correct password to validate")
	}
	if svc.ValidatePassword("otra-cosa", hash) {
		t.Error("expected a wrong password to fail")
	}
}
