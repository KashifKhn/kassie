package client

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewClient(t *testing.T) {
	c, err := New("localhost:0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer c.Close()

	if c.conn == nil {
		t.Fatal("expected conn to be set")
	}
	if c.session == nil {
		t.Fatal("expected session client to be set")
	}
	if c.schema == nil {
		t.Fatal("expected schema client to be set")
	}
	if c.data == nil {
		t.Fatal("expected data client to be set")
	}
}

func TestIsAuthenticated(t *testing.T) {
	c, err := New("localhost:0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer c.Close()

	if c.IsAuthenticated() {
		t.Fatal("expected not authenticated initially")
	}

	c.mu.Lock()
	c.accessToken = "test-token"
	c.mu.Unlock()

	if !c.IsAuthenticated() {
		t.Fatal("expected authenticated after setting token")
	}
}

func TestProfile(t *testing.T) {
	c, err := New("localhost:0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer c.Close()

	if c.Profile() != "" {
		t.Fatal("expected empty profile initially")
	}

	c.mu.Lock()
	c.profile = "local"
	c.mu.Unlock()

	if c.Profile() != "local" {
		t.Fatalf("expected profile 'local', got %q", c.Profile())
	}
}

func TestNeedsRefresh(t *testing.T) {
	c, err := New("localhost:0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer c.Close()

	if c.needsRefresh() {
		t.Fatal("should not need refresh without token")
	}
}

func TestRefreshNoToken(t *testing.T) {
	c, err := New("localhost:0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer c.Close()

	err = c.Refresh(context.Background())
	if err == nil {
		t.Fatal("expected error when no refresh token")
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{
			name:   "unauthenticated code",
			err:    status.Error(codes.Unauthenticated, "no token"),
			expect: true,
		},
		{
			name:   "token expired message",
			err:    status.Error(codes.Unknown, "Token expired"),
			expect: true,
		},
		{
			name:   "permission denied",
			err:    status.Error(codes.PermissionDenied, "forbidden"),
			expect: false,
		},
		{
			name:   "internal error",
			err:    status.Error(codes.Internal, "server error"),
			expect: false,
		},
		{
			name:   "nil error",
			err:    nil,
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAuthError(tt.err)
			if got != tt.expect {
				t.Errorf("isAuthError() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestPublicMethods(t *testing.T) {
	expected := []string{
		"/kassie.v1.SessionService/Login",
		"/kassie.v1.SessionService/Refresh",
		"/kassie.v1.SessionService/GetProfiles",
	}

	for _, m := range expected {
		if !publicMethods[m] {
			t.Errorf("expected %s to be public", m)
		}
	}

	private := []string{
		"/kassie.v1.SessionService/Logout",
		"/kassie.v1.SchemaService/ListKeyspaces",
		"/kassie.v1.DataService/QueryRows",
	}

	for _, m := range private {
		if publicMethods[m] {
			t.Errorf("expected %s to NOT be public", m)
		}
	}
}

func TestCloseNilConn(t *testing.T) {
	c := &Client{}
	err := c.Close()
	if err != nil {
		t.Fatalf("expected no error closing nil conn, got %v", err)
	}
}
