package grpc

import (
	"context"
	"strings"

	"github.com/KashifKhn/kassie/internal/server/service"
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

func NewAuthInterceptor(auth TokenValidator, store SessionStore, log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		ctx, err := authenticateContext(ctx, auth, store, log)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func NewStreamAuthInterceptor(auth TokenValidator, store SessionStore, log *logger.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if publicMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		ctx, err := authenticateContext(ss.Context(), auth, store, log)
		if err != nil {
			return err
		}

		return handler(srv, &authenticatedServerStream{ServerStream: ss, ctx: ctx})
	}
}

type authenticatedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *authenticatedServerStream) Context() context.Context {
	return s.ctx
}

func authenticateContext(ctx context.Context, auth TokenValidator, store SessionStore, log *logger.Logger) (context.Context, error) {
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

	return ctx, nil
}
