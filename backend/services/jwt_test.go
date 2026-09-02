package services

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	const secret = "test-secret"

	token, err := GenerateToken(42, "user@example.com", secret)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken() returned an empty token")
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("UserID = %d, want 42", claims.UserID)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("Email = %q, want user@example.com", claims.Email)
	}
	if claims.Issuer != "notell-api" {
		t.Fatalf("Issuer = %q, want notell-api", claims.Issuer)
	}
	if claims.Subject != "42" {
		t.Fatalf("Subject = %q, want 42", claims.Subject)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.After(time.Now()) {
		t.Fatal("token expiration was not set to a future time")
	}
}

func TestValidateTokenRejectsWrongSecret(t *testing.T) {
	token, err := GenerateToken(42, "user@example.com", "correct-secret")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if _, err := ValidateToken(token, "wrong-secret"); err == nil {
		t.Fatal("ValidateToken() accepted a token signed with the wrong secret")
	}
}

func TestValidateTokenRejectsWrongSigningMethod(t *testing.T) {
	now := time.Now()
	claims := Claims{
		UserID: 42,
		Email:  "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:   "notell-api",
			Subject:  "42",
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS384, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := ValidateToken(token, "test-secret"); err == nil {
		t.Fatal("ValidateToken() accepted a token using an unexpected signing method")
	}
}

func TestValidateTokenRejectsMissingToken(t *testing.T) {
	if _, err := ValidateToken("", "test-secret"); err == nil {
		t.Fatal("ValidateToken() accepted an empty token")
	}
}

func TestGeneratedTokenUsesExpectedStructure(t *testing.T) {
	token, err := GenerateToken(7, "seven@example.com", "test-secret")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if parts := strings.Split(token, "."); len(parts) != 3 {
		t.Fatalf("JWT segments = %d, want 3", len(parts))
	}
}
