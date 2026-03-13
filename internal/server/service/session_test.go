package service

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/db"
	"github.com/KashifKhn/kassie/internal/server/state"
	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/gocql/gocql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockPool struct {
	session *gocql.Session
	err     error
}

func (m *mockPool) GetOrCreate(profileName string, cfg *db.ConnectionConfig) (*gocql.Session, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.session, nil
}

type mockProfileProvider struct {
	profiles map[string]*config.Profile
}

func (m *mockProfileProvider) GetProfile(name string) (*config.Profile, error) {
	if p, ok := m.profiles[name]; ok {
		return p, nil
	}
	return nil, errors.New("profile not found")
}

func (m *mockProfileProvider) GetProfiles() []config.Profile {
	profiles := make([]config.Profile, 0, len(m.profiles))
	for _, p := range m.profiles {
		profiles = append(profiles, *p)
	}
	return profiles
}

type mockSessionStore struct {
	sessions map[string]*state.Session
}

func newMockSessionStore() *mockSessionStore {
	return &mockSessionStore{
		sessions: make(map[string]*state.Session),
	}
}

func (m *mockSessionStore) Create(id string, profile *config.Profile, conn *db.Session) *state.Session {
	sess := &state.Session{
		ID:         id,
		Connection: conn,
		Profile:    profile,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
		Cursors:    state.NewCursorStore(time.Hour),
	}
	m.sessions[id] = sess
	return sess
}

func (m *mockSessionStore) Get(id string) (*state.Session, error) {
	if sess, ok := m.sessions[id]; ok {
		return sess, nil
	}
	return nil, state.ErrSessionNotFound
}

func (m *mockSessionStore) Delete(id string) {
	delete(m.sessions, id)
}

func (m *mockSessionStore) CloseAll() {
	m.sessions = make(map[string]*state.Session)
}

func (m *mockSessionStore) Close() {
}

func TestSessionService_Login_MissingProfile(t *testing.T) {
	cfg := &mockProfileProvider{profiles: make(map[string]*config.Profile)}
	pool := &mockPool{}
	store := newMockSessionStore()
	auth := NewAuthService("test-secret")
	service := NewSessionService(cfg, pool, store, auth)

	_, err := service.Login(context.Background(), &pb.LoginRequest{Profile: ""})

	if err == nil {
		t.Fatal("expected error for missing profile")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestSessionService_Login_ProfileNotFound(t *testing.T) {
	cfg := &mockProfileProvider{profiles: make(map[string]*config.Profile)}
	pool := &mockPool{}
	store := newMockSessionStore()
	auth := NewAuthService("test-secret")
	service := NewSessionService(cfg, pool, store, auth)

	_, err := service.Login(context.Background(), &pb.LoginRequest{Profile: "nonexistent"})

	if err == nil {
		t.Fatal("expected error for non-existent profile")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}

	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}

func TestSessionService_Login_ConnectionFailed(t *testing.T) {
	cfg := &mockProfileProvider{
		profiles: map[string]*config.Profile{
			"test": {Name: "test", Hosts: []string{"localhost"}, Port: 9042},
		},
	}
	pool := &mockPool{err: errors.New("connection failed")}
	store := newMockSessionStore()
	auth := NewAuthService("test-secret")
	service := NewSessionService(cfg, pool, store, auth)

	_, err := service.Login(context.Background(), &pb.LoginRequest{Profile: "test"})

	if err == nil {
		t.Fatal("expected error for connection failure")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}

	if st.Code() != codes.Unavailable {
		t.Errorf("expected Unavailable, got %v", st.Code())
	}
}

func TestSessionService_Refresh_MissingToken(t *testing.T) {
	cfg := &mockProfileProvider{profiles: make(map[string]*config.Profile)}
	pool := &mockPool{}
	store := newMockSessionStore()
	auth := NewAuthService("test-secret")
	service := NewSessionService(cfg, pool, store, auth)

	_, err := service.Refresh(context.Background(), &pb.RefreshRequest{RefreshToken: ""})

	if err == nil {
		t.Fatal("expected error for missing refresh token")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestSessionService_Refresh_InvalidToken(t *testing.T) {
	cfg := &mockProfileProvider{profiles: make(map[string]*config.Profile)}
	pool := &mockPool{}
	store := newMockSessionStore()
	auth := NewAuthService("test-secret")
	service := NewSessionService(cfg, pool, store, auth)

	_, err := service.Refresh(context.Background(), &pb.RefreshRequest{RefreshToken: "invalid-token"})

	if err == nil {
		t.Fatal("expected error for invalid refresh token")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}

	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestSessionService_GetProfiles(t *testing.T) {
	cfg := &mockProfileProvider{
		profiles: map[string]*config.Profile{
			"local": {Name: "local", Hosts: []string{"localhost"}, Port: 9042},
			"prod":  {Name: "prod", Hosts: []string{"prod.example.com"}, Port: 9042},
		},
	}
	pool := &mockPool{}
	store := newMockSessionStore()
	auth := NewAuthService("test-secret")
	service := NewSessionService(cfg, pool, store, auth)

	resp, err := service.GetProfiles(context.Background(), &pb.GetProfilesRequest{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(resp.Profiles))
	}
}

func TestGetSessionFromContext_NoSessionID(t *testing.T) {
	store := newMockSessionStore()

	_, err := GetSessionFromContext(context.Background(), store)

	if err == nil {
		t.Fatal("expected error for missing session ID")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}

	if st.Code() != codes.Internal {
		t.Errorf("expected Internal, got %v", st.Code())
	}
}
