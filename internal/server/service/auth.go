package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrInvalidClaims    = errors.New("invalid token claims")
	ErrMissingSessionID = errors.New("missing session id")
	ErrMissingProfile   = errors.New("missing profile")
	ErrInvalidSecret    = errors.New("invalid secret key")
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	SessionID string    `json:"session_id"`
	Profile   string    `json:"profile"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type AuthService struct {
	secret []byte
}

func NewAuthService(secret string) (*AuthService, error) {
	if secret == "" {
		return nil, ErrInvalidSecret
	}
	return &AuthService{
		secret: []byte(secret),
	}, nil
}

func (a *AuthService) GenerateAccessToken(sessionID, profile string) (string, time.Time, error) {
	if sessionID == "" {
		return "", time.Time{}, ErrMissingSessionID
	}
	if profile == "" {
		return "", time.Time{}, ErrMissingProfile
	}

	expiresAt := time.Now().Add(AccessTokenDuration)
	claims := Claims{
		SessionID: sessionID,
		Profile:   profile,
		TokenType: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(a.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, expiresAt, nil
}

func (a *AuthService) GenerateRefreshToken(sessionID string) (string, time.Time, error) {
	if sessionID == "" {
		return "", time.Time{}, ErrMissingSessionID
	}

	expiresAt := time.Now().Add(RefreshTokenDuration)
	claims := Claims{
		SessionID: sessionID,
		TokenType: RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(a.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, expiresAt, nil
}

func (a *AuthService) ValidateToken(tokenString string, expectedType TokenType) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("%w: expected %s, got %s", ErrInvalidClaims, expectedType, claims.TokenType)
	}

	if claims.SessionID == "" {
		return nil, ErrMissingSessionID
	}

	return claims, nil
}

func (a *AuthService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return a.ValidateToken(tokenString, AccessToken)
}

func (a *AuthService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return a.ValidateToken(tokenString, RefreshToken)
}

func (a *AuthService) ExtractSessionID(tokenString string) (string, error) {
	claims, err := a.ValidateAccessToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.SessionID, nil
}
