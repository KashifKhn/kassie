package grpc

import (
	"context"
	"strings"

	"github.com/KashifKhn/kassie/internal/server/service"
	"github.com/KashifKhn/kassie/internal/server/state"
	"github.com/KashifKhn/kassie/internal/shared/ctxutil"
	"github.com/KashifKhn/kassie/internal/shared/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var publicMethods = map[string]bool{
	"/kassie.v1.SessionService/Login":       true,
	"/kassie.v1.SessionService/Refresh":     true,
	"/kassie.v1.SessionService/GetProfiles": true,
}

func NewAuthInterceptor(auth *service.AuthService, store *state.Store, log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Warn("no metadata in request")
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			log.Warn("no authorization header")
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		if token == authHeader[0] {
			log.Warn("invalid authorization format")
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}

		claims, err := auth.ValidateToken(token, service.AccessToken)
		if err != nil {
			log.With().Err(err).Logger().Warn("token validation failed")
			if err == service.ErrExpiredToken {
				return nil, status.Error(codes.Unauthenticated, "token expired")
			}
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		session, err := store.Get(claims.SessionID)
		if err != nil {
			log.With().Str("session_id", claims.SessionID).Err(err).Logger().Warn("session not found")
			return nil, status.Error(codes.Unauthenticated, "session not found or expired")
		}

		ctx = ctxutil.WithSessionID(ctx, session.ID)
		ctx = ctxutil.WithProfile(ctx, claims.Profile)

		return handler(ctx, req)
	}
}
