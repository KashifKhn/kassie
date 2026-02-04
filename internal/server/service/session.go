package service

import (
	"context"
	"fmt"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/db"
	"github.com/KashifKhn/kassie/internal/server/state"
	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/google/uuid"
)

type SessionService struct {
	pb.UnimplementedSessionServiceServer
	cfg   *config.Config
	pool  *db.Pool
	store *state.Store
	auth  *AuthService
}

func NewSessionService(cfg *config.Config, pool *db.Pool, store *state.Store, auth *AuthService) *SessionService {
	return &SessionService{
		cfg:   cfg,
		pool:  pool,
		store: store,
		auth:  auth,
	}
}

func (s *SessionService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Profile == "" {
		return nil, fmt.Errorf("profile name is required")
	}

	profile, err := s.cfg.GetProfile(req.Profile)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %s", req.Profile)
	}

	connCfg := db.ProfileToConnectionConfig(profile)
	gocqlSession, err := s.pool.GetOrCreate(profile.Name, connCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sessionID := uuid.New().String()
	dbSession := db.NewSession(gocqlSession)
	s.store.Create(sessionID, profile, dbSession)

	accessToken, refreshToken, expiresAt, err := s.auth.GenerateTokenPair(sessionID, profile.Name)
	if err != nil {
		s.store.Delete(sessionID)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
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
		return nil, fmt.Errorf("refresh token is required")
	}

	accessToken, expiresAt, err := s.auth.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &pb.RefreshResponse{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *SessionService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	sessionID := ctx.Value("session_id")
	if sessionID != nil {
		if sid, ok := sessionID.(string); ok {
			s.store.Delete(sid)
		}
	}

	return &pb.LogoutResponse{}, nil
}

func (s *SessionService) GetProfiles(ctx context.Context, req *pb.GetProfilesRequest) (*pb.GetProfilesResponse, error) {
	profiles := make([]*pb.ProfileInfo, 0, len(s.cfg.Profiles))

	for i := range s.cfg.Profiles {
		p := &s.cfg.Profiles[i]
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

func GetSessionFromContext(ctx context.Context, store *state.Store) (*state.Session, error) {
	sessionID := ctx.Value("session_id")
	if sessionID == nil {
		return nil, fmt.Errorf("no session in context")
	}

	sid, ok := sessionID.(string)
	if !ok {
		return nil, fmt.Errorf("invalid session ID type")
	}

	session, err := store.Get(sid)
	if err != nil {
		return nil, fmt.Errorf("session not found or expired: %w", err)
	}

	session.LastAccess = time.Now()
	return session, nil
}
