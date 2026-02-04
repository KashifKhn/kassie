package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token expired")
	ErrInvalidClaims = errors.New("invalid token claims")
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	SessionID string    `json:"session_id"`
	Profile   string    `json:"profile"`
	Type      TokenType `json:"type"`
	jwt.RegisteredClaims
}

type AuthService struct {
	secretKey []byte
}

func NewAuthService(secretKey string) *AuthService {
	return &AuthService{
		secretKey: []byte(secretKey),
	}
}

func (a *AuthService) GenerateTokenPair(sessionID, profile string) (accessToken, refreshToken string, expiresAt int64, err error) {
	now := time.Now()

	accessClaims := &Claims{
		SessionID: sessionID,
		Profile:   profile,
		Type:      AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	accessTkn := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTkn.SignedString(a.secretKey)
	if err != nil {
		return "", "", 0, err
	}

	refreshClaims := &Claims{
		SessionID: sessionID,
		Profile:   profile,
		Type:      RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	refreshTkn := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTkn.SignedString(a.secretKey)
	if err != nil {
		return "", "", 0, err
	}

	expiresAt = accessClaims.ExpiresAt.Unix()
	return accessToken, refreshToken, expiresAt, nil
}

func (a *AuthService) ValidateToken(tokenString string, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return a.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	if claims.Type != expectedType {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

func (a *AuthService) RefreshAccessToken(refreshTokenString string) (accessToken string, expiresAt int64, err error) {
	claims, err := a.ValidateToken(refreshTokenString, RefreshToken)
	if err != nil {
		return "", 0, err
	}

	now := time.Now()
	accessClaims := &Claims{
		SessionID: claims.SessionID,
		Profile:   claims.Profile,
		Type:      AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = token.SignedString(a.secretKey)
	if err != nil {
		return "", 0, err
	}

	expiresAt = accessClaims.ExpiresAt.Unix()
	return accessToken, expiresAt, nil
}
