package service

import (
	"context"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/db"
	"github.com/KashifKhn/kassie/internal/server/state"
	"github.com/KashifKhn/kassie/internal/shared/ctxutil"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SessionService struct {
	pb.UnimplementedSessionServiceServer
	cfg   ProfileProvider
	pool  ConnectionPool
	store SessionStore
	auth  *AuthService
}

func NewSessionService(cfg ProfileProvider, pool ConnectionPool, store SessionStore, auth *AuthService) *SessionService {
	return &SessionService{
		cfg:   cfg,
		pool:  pool,
		store: store,
		auth:  auth,
	}
}

func (s *SessionService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Profile == "" {
		return nil, status.Error(codes.InvalidArgument, "profile name is required")
	}

	profile, err := s.cfg.GetProfile(req.Profile)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "profile not found: %s", req.Profile)
	}

	connCfg := db.ProfileToConnectionConfig(profile)
	gocqlSession, err := s.pool.GetOrCreate(profile.Name, connCfg)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "failed to connect to database: %v", err)
	}

	sessionID := uuid.New().String()
	dbSession := db.NewSession(gocqlSession)
	s.store.Create(sessionID, profile, dbSession)

	accessToken, refreshToken, expiresAt, err := s.auth.GenerateTokenPair(sessionID, profile.Name)
	if err != nil {
		s.store.Delete(sessionID)
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Profile: &pb.ProfileInfo{
			Name:       profile.Name,
			Hosts:      profile.Hosts,
			Port:       int32(profile.Port),
			Keyspace:   profile.Keyspace,
			SslEnabled: profile.SSL != nil && profile.SSL.Enabled,
		},
	}, nil
}

func (s *SessionService) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	accessToken, expiresAt, err := s.auth.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to refresh token: %v", err)
	}

	return &pb.RefreshResponse{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *SessionService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if sessionID, ok := ctxutil.GetSessionID(ctx); ok {
		s.store.Delete(sessionID)
	}

	return &pb.LogoutResponse{}, nil
}

func (s *SessionService) GetProfiles(ctx context.Context, req *pb.GetProfilesRequest) (*pb.GetProfilesResponse, error) {
	profileList := s.cfg.GetProfiles()
	profiles := make([]*pb.ProfileInfo, 0, len(profileList))

	for i := range profileList {
		p := &profileList[i]
		profiles = append(profiles, &pb.ProfileInfo{
			Name:       p.Name,
			Hosts:      p.Hosts,
			Port:       int32(p.Port),
			Keyspace:   p.Keyspace,
			SslEnabled: p.SSL != nil && p.SSL.Enabled,
		})
	}

	return &pb.GetProfilesResponse{
		Profiles: profiles,
	}, nil
}

func GetSessionFromContext(ctx context.Context, store SessionStore) (*state.Session, error) {
	sessionID, ok := ctxutil.GetSessionID(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "no session in context")
	}

	session, err := store.Get(sessionID)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "session not found or expired: %v", err)
	}

	session.LastAccess = time.Now()
	return session, nil
}
