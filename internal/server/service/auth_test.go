package service

import (
	"testing"
	"time"
)

func TestAuthService_GenerateTokenPair(t *testing.T) {
	auth := NewAuthService("test-secret-key")

	accessToken, refreshToken, expiresAt, err := auth.GenerateTokenPair("session-1", "test-profile")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if accessToken == "" {
		t.Error("expected non-empty access token")
	}

	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}

	if expiresAt == 0 {
		t.Error("expected non-zero expires at")
	}

	now := time.Now().Unix()
	expectedExpiry := now + int64(AccessTokenDuration.Seconds())
	if expiresAt < now || expiresAt > expectedExpiry+10 {
		t.Errorf("expires at %d is out of expected range [%d, %d]", expiresAt, now, expectedExpiry+10)
	}
}

func TestAuthService_ValidateAccessToken(t *testing.T) {
	auth := NewAuthService("test-secret-key")

	accessToken, _, _, err := auth.GenerateTokenPair("session-1", "test-profile")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := auth.ValidateToken(accessToken, AccessToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if claims.SessionID != "session-1" {
		t.Errorf("expected session ID session-1, got %s", claims.SessionID)
	}

	if claims.Profile != "test-profile" {
		t.Errorf("expected profile test-profile, got %s", claims.Profile)
	}

	if claims.Type != AccessToken {
		t.Errorf("expected token type access, got %s", claims.Type)
	}
}

func TestAuthService_ValidateRefreshToken(t *testing.T) {
	auth := NewAuthService("test-secret-key")

	_, refreshToken, _, err := auth.GenerateTokenPair("session-1", "test-profile")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := auth.ValidateToken(refreshToken, RefreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if claims.SessionID != "session-1" {
		t.Errorf("expected session ID session-1, got %s", claims.SessionID)
	}

	if claims.Type != RefreshToken {
		t.Errorf("expected token type refresh, got %s", claims.Type)
	}
}

func TestAuthService_ValidateToken_WrongType(t *testing.T) {
	auth := NewAuthService("test-secret-key")

	accessToken, _, _, err := auth.GenerateTokenPair("session-1", "test-profile")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = auth.ValidateToken(accessToken, RefreshToken)
	if err != ErrInvalidClaims {
		t.Errorf("expected ErrInvalidClaims, got %v", err)
	}
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	auth := NewAuthService("test-secret-key")

	_, err := auth.ValidateToken("invalid-token", AccessToken)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthService_ValidateToken_WrongSecret(t *testing.T) {
	auth1 := NewAuthService("secret-1")
	auth2 := NewAuthService("secret-2")

	token, _, _, err := auth1.GenerateTokenPair("session-1", "test-profile")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = auth2.ValidateToken(token, AccessToken)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for wrong secret, got %v", err)
	}
}

func TestAuthService_RefreshAccessToken(t *testing.T) {
	auth := NewAuthService("test-secret-key")

	_, refreshToken, _, err := auth.GenerateTokenPair("session-1", "test-profile")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	newAccessToken, expiresAt, err := auth.RefreshAccessToken(refreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if newAccessToken == "" {
		t.Error("expected non-empty new access token")
	}

	if expiresAt == 0 {
		t.Error("expected non-zero expires at")
	}

	claims, err := auth.ValidateToken(newAccessToken, AccessToken)
	if err != nil {
		t.Fatalf("failed to validate new access token: %v", err)
	}

	if claims.SessionID != "session-1" {
		t.Errorf("expected session ID session-1, got %s", claims.SessionID)
	}
}

func TestAuthService_RefreshAccessToken_InvalidRefreshToken(t *testing.T) {
	auth := NewAuthService("test-secret-key")

	_, _, err := auth.RefreshAccessToken("invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthService_RefreshAccessToken_WithAccessToken(t *testing.T) {
	auth := NewAuthService("test-secret-key")

	accessToken, _, _, err := auth.GenerateTokenPair("session-1", "test-profile")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, _, err = auth.RefreshAccessToken(accessToken)
	if err != ErrInvalidClaims {
		t.Errorf("expected ErrInvalidClaims when using access token, got %v", err)
	}
}
