package service

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewAuthService(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		wantErr   bool
		expectErr error
	}{
		{
			name:      "valid secret",
			secret:    "test-secret-key",
			wantErr:   false,
			expectErr: nil,
		},
		{
			name:      "empty secret",
			secret:    "",
			wantErr:   true,
			expectErr: ErrInvalidSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewAuthService(tt.secret)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if err != tt.expectErr {
					t.Errorf("expected error %v, got %v", tt.expectErr, err)
				}
				if svc != nil {
					t.Errorf("expected nil service, got %v", svc)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if svc == nil {
					t.Error("expected service, got nil")
				}
			}
		})
	}
}

func TestGenerateAccessToken(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	tests := []struct {
		name      string
		sessionID string
		profile   string
		wantErr   bool
		expectErr error
	}{
		{
			name:      "valid inputs",
			sessionID: "session-123",
			profile:   "prod",
			wantErr:   false,
		},
		{
			name:      "empty session id",
			sessionID: "",
			profile:   "prod",
			wantErr:   true,
			expectErr: ErrMissingSessionID,
		},
		{
			name:      "empty profile",
			sessionID: "session-123",
			profile:   "",
			wantErr:   true,
			expectErr: ErrMissingProfile,
		},
		{
			name:      "both empty",
			sessionID: "",
			profile:   "",
			wantErr:   true,
			expectErr: ErrMissingSessionID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiresAt, err := svc.GenerateAccessToken(tt.sessionID, tt.profile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if err != tt.expectErr {
					t.Errorf("expected error %v, got %v", tt.expectErr, err)
				}
				if token != "" {
					t.Errorf("expected empty token, got %s", token)
				}
				if !expiresAt.IsZero() {
					t.Errorf("expected zero time, got %v", expiresAt)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if token == "" {
					t.Error("expected non-empty token")
				}
				if expiresAt.IsZero() {
					t.Error("expected non-zero expiry time")
				}

				expectedExpiry := time.Now().Add(AccessTokenDuration)
				diff := expiresAt.Sub(expectedExpiry)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("expiry time off by %v", diff)
				}
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
		expectErr error
	}{
		{
			name:      "valid session id",
			sessionID: "session-456",
			wantErr:   false,
		},
		{
			name:      "empty session id",
			sessionID: "",
			wantErr:   true,
			expectErr: ErrMissingSessionID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiresAt, err := svc.GenerateRefreshToken(tt.sessionID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if err != tt.expectErr {
					t.Errorf("expected error %v, got %v", tt.expectErr, err)
				}
				if token != "" {
					t.Errorf("expected empty token, got %s", token)
				}
				if !expiresAt.IsZero() {
					t.Errorf("expected zero time, got %v", expiresAt)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if token == "" {
					t.Error("expected non-empty token")
				}
				if expiresAt.IsZero() {
					t.Error("expected non-zero expiry time")
				}

				expectedExpiry := time.Now().Add(RefreshTokenDuration)
				diff := expiresAt.Sub(expectedExpiry)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("expiry time off by %v", diff)
				}
			}
		})
	}
}

func TestValidateAccessToken(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	validToken, _, err := svc.GenerateAccessToken("session-789", "dev")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	wrongTypeSvc, _ := NewAuthService("test-secret")
	refreshToken, _, _ := wrongTypeSvc.GenerateRefreshToken("session-789")

	wrongSecretSvc, _ := NewAuthService("wrong-secret")
	wrongSecretToken, _, _ := wrongSecretSvc.GenerateAccessToken("session-789", "dev")

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		expectErr error
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:      "empty token",
			token:     "",
			wantErr:   true,
			expectErr: ErrInvalidToken,
		},
		{
			name:      "malformed token",
			token:     "not.a.jwt",
			wantErr:   true,
			expectErr: ErrInvalidToken,
		},
		{
			name:      "wrong token type",
			token:     refreshToken,
			wantErr:   true,
			expectErr: ErrInvalidClaims,
		},
		{
			name:      "wrong secret",
			token:     wrongSecretToken,
			wantErr:   true,
			expectErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := svc.ValidateAccessToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if claims != nil {
					t.Errorf("expected nil claims, got %v", claims)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if claims == nil {
					t.Error("expected claims, got nil")
				}
				if claims.SessionID != "session-789" {
					t.Errorf("expected session-789, got %s", claims.SessionID)
				}
				if claims.Profile != "dev" {
					t.Errorf("expected dev profile, got %s", claims.Profile)
				}
				if claims.TokenType != AccessToken {
					t.Errorf("expected access token type, got %s", claims.TokenType)
				}
			}
		})
	}
}

func TestValidateRefreshToken(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	validToken, _, err := svc.GenerateRefreshToken("session-999")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	accessToken, _, _ := svc.GenerateAccessToken("session-999", "prod")

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		expectErr error
	}{
		{
			name:    "valid refresh token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:      "empty token",
			token:     "",
			wantErr:   true,
			expectErr: ErrInvalidToken,
		},
		{
			name:      "wrong token type (access instead of refresh)",
			token:     accessToken,
			wantErr:   true,
			expectErr: ErrInvalidClaims,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := svc.ValidateRefreshToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if claims != nil {
					t.Errorf("expected nil claims, got %v", claims)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if claims == nil {
					t.Error("expected claims, got nil")
				}
				if claims.SessionID != "session-999" {
					t.Errorf("expected session-999, got %s", claims.SessionID)
				}
				if claims.TokenType != RefreshToken {
					t.Errorf("expected refresh token type, got %s", claims.TokenType)
				}
			}
		})
	}
}

func TestExpiredToken(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	pastTime := time.Now().Add(-1 * time.Hour)
	claims := Claims{
		SessionID: "expired-session",
		Profile:   "test",
		TokenType: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(pastTime),
			IssuedAt:  jwt.NewNumericDate(pastTime.Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, err := token.SignedString(svc.secret)
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}

	_, err = svc.ValidateAccessToken(expiredToken)
	if err == nil {
		t.Error("expected error for expired token")
	}
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestExtractSessionID(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	token, _, err := svc.GenerateAccessToken("extracted-session", "staging")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		want      string
		wantErr   bool
		expectErr error
	}{
		{
			name:    "valid token",
			token:   token,
			want:    "extracted-session",
			wantErr: false,
		},
		{
			name:      "empty token",
			token:     "",
			want:      "",
			wantErr:   true,
			expectErr: ErrInvalidToken,
		},
		{
			name:      "invalid token",
			token:     "invalid.token.here",
			want:      "",
			wantErr:   true,
			expectErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID, err := svc.ExtractSessionID(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if sessionID != "" {
					t.Errorf("expected empty session id, got %s", sessionID)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if sessionID != tt.want {
					t.Errorf("expected %s, got %s", tt.want, sessionID)
				}
			}
		})
	}
}

func TestTokenSigningMethod(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	claims := Claims{
		SessionID: "test-session",
		Profile:   "test",
		TokenType: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	rsaToken, err := token.SignedString(svc.secret)

	if err == nil {
		_, err = svc.ValidateAccessToken(rsaToken)
		if err == nil {
			t.Error("expected error for wrong signing method")
		}
		if !strings.Contains(err.Error(), "unexpected signing method") && err != ErrInvalidToken {
			t.Errorf("expected signing method error, got %v", err)
		}
	}
}

func TestTokenWithoutSessionID(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	claims := Claims{
		SessionID: "",
		Profile:   "test",
		TokenType: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	noSessionToken, err := token.SignedString(svc.secret)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	_, err = svc.ValidateAccessToken(noSessionToken)
	if err == nil {
		t.Error("expected error for missing session id")
	}
	if err != ErrMissingSessionID {
		t.Errorf("expected ErrMissingSessionID, got %v", err)
	}
}

func TestValidateTokenGeneric(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	accessToken, _, _ := svc.GenerateAccessToken("session-1", "profile-1")
	refreshToken, _, _ := svc.GenerateRefreshToken("session-2")

	tests := []struct {
		name         string
		token        string
		expectedType TokenType
		wantErr      bool
	}{
		{
			name:         "validate access as access",
			token:        accessToken,
			expectedType: AccessToken,
			wantErr:      false,
		},
		{
			name:         "validate refresh as refresh",
			token:        refreshToken,
			expectedType: RefreshToken,
			wantErr:      false,
		},
		{
			name:         "validate access as refresh",
			token:        accessToken,
			expectedType: RefreshToken,
			wantErr:      true,
		},
		{
			name:         "validate refresh as access",
			token:        refreshToken,
			expectedType: AccessToken,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := svc.ValidateToken(tt.token, tt.expectedType)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if claims == nil {
					t.Error("expected claims, got nil")
				}
				if claims.TokenType != tt.expectedType {
					t.Errorf("expected type %s, got %s", tt.expectedType, claims.TokenType)
				}
			}
		})
	}
}

func TestConcurrentTokenGeneration(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	done := make(chan bool)
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			token, _, err := svc.GenerateAccessToken("session-concurrent", "profile")
			if err != nil {
				errors <- err
			} else if token == "" {
				errors <- ErrInvalidToken
			}
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
	close(errors)

	for err := range errors {
		t.Errorf("concurrent generation failed: %v", err)
	}
}

func TestInvalidClaimsStructure(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	type BadClaims struct {
		jwt.RegisteredClaims
	}

	claims := BadClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	badToken, err := token.SignedString(svc.secret)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	_, err = svc.ValidateAccessToken(badToken)
	if err == nil {
		t.Error("expected error for invalid claims structure")
	}
}

func TestTokenDurationConstants(t *testing.T) {
	if AccessTokenDuration != 15*time.Minute {
		t.Errorf("expected access token duration 15m, got %v", AccessTokenDuration)
	}
	if RefreshTokenDuration != 7*24*time.Hour {
		t.Errorf("expected refresh token duration 7d, got %v", RefreshTokenDuration)
	}
}

func TestTokenTypeConstants(t *testing.T) {
	if AccessToken != "access" {
		t.Errorf("expected access token type 'access', got %s", AccessToken)
	}
	if RefreshToken != "refresh" {
		t.Errorf("expected refresh token type 'refresh', got %s", RefreshToken)
	}
}

func TestValidateTokenWithInvalidAlgorithm(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	token, _, err := svc.GenerateAccessToken("test", "profile")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("invalid token format")
	}

	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if claims == nil {
		t.Error("expected claims")
	}
}

func TestValidateWithEmptySecret(t *testing.T) {
	_, err := NewAuthService("")
	if err != ErrInvalidSecret {
		t.Errorf("expected ErrInvalidSecret, got %v", err)
	}
}

func TestGenerateTokensVerifyExpiryAccuracy(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	beforeAccess := time.Now().Add(AccessTokenDuration)
	accessToken, accessExpiry, err := svc.GenerateAccessToken("test", "profile")
	afterAccess := time.Now().Add(AccessTokenDuration)

	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	if accessToken == "" {
		t.Error("access token is empty")
	}

	if accessExpiry.Before(beforeAccess.Add(-time.Second)) || accessExpiry.After(afterAccess.Add(time.Second)) {
		t.Errorf("access expiry time is not within expected range")
	}

	beforeRefresh := time.Now().Add(RefreshTokenDuration)
	refreshToken, refreshExpiry, err := svc.GenerateRefreshToken("test")
	afterRefresh := time.Now().Add(RefreshTokenDuration)

	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	if refreshToken == "" {
		t.Error("refresh token is empty")
	}

	if refreshExpiry.Before(beforeRefresh.Add(-time.Second)) || refreshExpiry.After(afterRefresh.Add(time.Second)) {
		t.Errorf("refresh expiry time is not within expected range")
	}
}

func TestValidateTokenErrorPaths(t *testing.T) {
	svc, err := NewAuthService("test-secret")
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "completely invalid token",
			token:   "this-is-not-a-jwt",
			wantErr: true,
		},
		{
			name:    "token with invalid parts",
			token:   "header.payload",
			wantErr: true,
		},
		{
			name:    "token with extra parts",
			token:   "a.b.c.d",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ValidateAccessToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got error=%v", tt.wantErr, err)
			}
		})
	}
}
