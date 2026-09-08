package state

import (
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *NLExampleStore {
	t.Helper()
	return NewNLExampleStore(filepath.Join(t.TempDir(), "nl-examples.json"))
}

func TestAddAndList(t *testing.T) {
	store := newTestStore(t)

	id := store.Add("prod", "critical events yesterday", "SELECT * FROM events", "app")
	if id == "" {
		t.Fatal("expected non-empty id")
	}

	got := store.List("prod", "", 0)
	if len(got) != 1 {
		t.Fatalf("expected 1 example, got %d", len(got))
	}
	if got[0].Question != "critical events yesterday" || got[0].CQL != "SELECT * FROM events" {
		t.Fatalf("unexpected example: %+v", got[0])
	}
}

func TestListFiltersByKeyspace(t *testing.T) {
	store := newTestStore(t)
	store.Add("prod", "q1", "SELECT 1", "app")
	store.Add("prod", "q2", "SELECT 2", "other")

	got := store.List("prod", "app", 0)
	if len(got) != 1 {
		t.Fatalf("expected 1 example for keyspace app, got %d", len(got))
	}
	if got[0].Question != "q1" {
		t.Fatalf("expected q1, got %s", got[0].Question)
	}
}

func TestListNewestFirst(t *testing.T) {
	store := newTestStore(t)
	store.Add("prod", "first", "SELECT 1", "app")
	store.Add("prod", "second", "SELECT 2", "app")

	got := store.List("prod", "", 0)
	if len(got) != 2 || got[0].Question != "second" {
		t.Fatalf("expected newest first, got %+v", got)
	}
}

func TestListLimit(t *testing.T) {
	store := newTestStore(t)
	for i := 0; i < 10; i++ {
		store.Add("prod", "q", "SELECT 1", "app")
	}
	if got := store.List("prod", "", 5); len(got) != 5 {
		t.Fatalf("expected limit 5, got %d", len(got))
	}
}

func TestAddCapsAtMax(t *testing.T) {
	store := newTestStore(t)
	for i := 0; i < MaxNLExamples+20; i++ {
		store.Add("prod", "q", "SELECT 1", "app")
	}
	if got := store.List("prod", "", 0); len(got) != MaxNLExamples {
		t.Fatalf("expected cap at %d, got %d", MaxNLExamples, len(got))
	}
}

func TestDelete(t *testing.T) {
	store := newTestStore(t)
	id := store.Add("prod", "q", "SELECT 1", "app")

	if !store.Delete("prod", id) {
		t.Fatal("expected delete to succeed")
	}
	if store.Delete("prod", id) {
		t.Fatal("expected second delete to fail")
	}
	if got := store.List("prod", "", 0); len(got) != 0 {
		t.Fatalf("expected empty list, got %d", len(got))
	}
}

func TestPersistRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nl-examples.json")
	store := NewNLExampleStore(path)
	store.Add("prod", "q", "SELECT 1", "app")
	if err := store.Persist(); err != nil {
		t.Fatal(err)
	}

	loaded := NewNLExampleStore(path)
	got := loaded.List("prod", "", 0)
	if len(got) != 1 || got[0].CQL != "SELECT 1" {
		t.Fatalf("expected round trip, got %+v", got)
	}
}

func TestEmptyAddIgnored(t *testing.T) {
	store := newTestStore(t)
	if store.Add("prod", "", "SELECT 1", "app") != "" {
		t.Fatal("expected empty question to be ignored")
	}
	if store.Add("prod", "q", "", "app") != "" {
		t.Fatal("expected empty cql to be ignored")
	}
}
