package grpc

import (
	"github.com/KashifKhn/kassie/internal/server/service"
	"github.com/KashifKhn/kassie/internal/server/state"
)

type TokenValidator interface {
	ValidateToken(tokenString string, expectedType service.TokenType) (*service.Claims, error)
}

type SessionStore interface {
	Get(id string) (*state.Session, error)
}
