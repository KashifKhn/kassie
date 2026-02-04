package db

import (
	"testing"

	"github.com/gocql/gocql"
)

func TestNewSession(t *testing.T) {
	mockSession := &gocql.Session{}
	session := NewSession(mockSession)

	if session == nil {
		t.Fatal("NewSession() returned nil")
	}

	if session.session != mockSession {
		t.Error("NewSession() did not set session correctly")
	}
}

func TestSessionClosed(t *testing.T) {
	mockSession := &gocql.Session{}
	session := NewSession(mockSession)

	if session.Closed() != mockSession.Closed() {
		t.Error("Closed() should return underlying session closed state")
	}
}
