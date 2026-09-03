package state

import (
	"testing"
	"time"

	"github.com/KashifKhn/kassie/internal/shared/config"
)

func TestStore_CreateAndGet(t *testing.T) {
	store := NewStore(30 * time.Minute)
	defer store.Close()

	profile := &config.Profile{
		Name:  "test",
		Hosts: []string{"localhost"},
		Port:  9042,
	}

	session := store.Create("session-1", profile, nil)

	if session.ID != "session-1" {
		t.Errorf("expected session ID session-1, got %s", session.ID)
	}

	if session.Profile.Name != "test" {
		t.Errorf("expected profile name test, got %s", session.Profile.Name)
	}

	retrieved, err := store.Get("session-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.ID != "session-1" {
		t.Errorf("expected session ID session-1, got %s", retrieved.ID)
	}
}

func TestStore_GetNonExistent(t *testing.T) {
	store := NewStore(30 * time.Minute)
	defer store.Close()

	_, err := store.Get("non-existent")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestStore_Delete(t *testing.T) {
	store := NewStore(30 * time.Minute)
	defer store.Close()

	profile := &config.Profile{
		Name:  "test",
		Hosts: []string{"localhost"},
		Port:  9042,
	}

	store.Create("session-1", profile, nil)

	if store.Count() != 1 {
		t.Errorf("expected count 1, got %d", store.Count())
	}

	store.Delete("session-1")

	if store.Count() != 0 {
		t.Errorf("expected count 0, got %d", store.Count())
	}

	_, err := store.Get("session-1")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound after delete, got %v", err)
	}
}

func TestStore_Expiry(t *testing.T) {
	store := NewStore(100 * time.Millisecond)
	defer store.Close()

	profile := &config.Profile{
		Name:  "test",
		Hosts: []string{"localhost"},
		Port:  9042,
	}

	store.Create("session-1", profile, nil)

	time.Sleep(200 * time.Millisecond)

	_, err := store.Get("session-1")
	if err != ErrSessionExpired {
		t.Errorf("expected ErrSessionExpired, got %v", err)
	}
}

func TestStore_CloseAll(t *testing.T) {
	store := NewStore(30 * time.Minute)
	defer store.Close()

	profile := &config.Profile{
		Name:  "test",
		Hosts: []string{"localhost"},
		Port:  9042,
	}

	store.Create("session-1", profile, nil)
	store.Create("session-2", profile, nil)

	if store.Count() != 2 {
		t.Errorf("expected count 2, got %d", store.Count())
	}

	store.CloseAll()

	if store.Count() != 0 {
		t.Errorf("expected count 0 after CloseAll, got %d", store.Count())
	}
}

func TestStoreTotalCursors(t *testing.T) {
	store := NewStore(time.Hour)
	defer store.CloseAll()

	p1 := &config.Profile{Name: "a", Hosts: []string{"h"}, Port: 9042}
	p2 := &config.Profile{Name: "b", Hosts: []string{"h"}, Port: 9042}

	s1 := store.Create("sess1", p1, nil)
	store.Create("sess2", p2, nil)

	if got := store.TotalCursors(); got != 0 {
		t.Fatalf("initial cursors = %d, want 0", got)
	}

	s1.Cursors.Create([]byte("ps1"), "ks", "tbl", "", 10)
	s1.Cursors.Create([]byte("ps2"), "ks", "tbl", "", 10)

	if got := store.Count(); got != 2 {
		t.Errorf("sessions = %d, want 2", got)
	}
	if got := store.TotalCursors(); got != 2 {
		t.Errorf("cursors = %d, want 2", got)
	}
}
