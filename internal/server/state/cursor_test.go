package state

import (
	"testing"
	"time"
)

func TestCursorStore_CreateAndGet(t *testing.T) {
	store := NewCursorStore(30 * time.Minute)

	pageState := []byte("test-page-state")
	cursorID := store.Create(pageState, "keyspace1", "table1", "", 100)

	if cursorID == "" {
		t.Fatal("expected non-empty cursor ID")
	}

	cursor, err := store.Get(cursorID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cursor.Keyspace != "keyspace1" {
		t.Errorf("expected keyspace keyspace1, got %s", cursor.Keyspace)
	}

	if cursor.Table != "table1" {
		t.Errorf("expected table table1, got %s", cursor.Table)
	}

	if cursor.PageSize != 100 {
		t.Errorf("expected page size 100, got %d", cursor.PageSize)
	}

	if string(cursor.PageState) != "test-page-state" {
		t.Errorf("expected page state test-page-state, got %s", string(cursor.PageState))
	}
}

func TestCursorStore_GetNonExistent(t *testing.T) {
	store := NewCursorStore(30 * time.Minute)

	_, err := store.Get("non-existent")
	if err != ErrCursorNotFound {
		t.Errorf("expected ErrCursorNotFound, got %v", err)
	}
}

func TestCursorStore_Delete(t *testing.T) {
	store := NewCursorStore(30 * time.Minute)

	cursorID := store.Create([]byte("state"), "ks", "tbl", "", 100)

	if store.Count() != 1 {
		t.Errorf("expected count 1, got %d", store.Count())
	}

	store.Delete(cursorID)

	if store.Count() != 0 {
		t.Errorf("expected count 0, got %d", store.Count())
	}

	_, err := store.Get(cursorID)
	if err != ErrCursorNotFound {
		t.Errorf("expected ErrCursorNotFound after delete, got %v", err)
	}
}

func TestCursorStore_Expiry(t *testing.T) {
	store := NewCursorStore(100 * time.Millisecond)

	cursorID := store.Create([]byte("state"), "ks", "tbl", "", 100)

	time.Sleep(200 * time.Millisecond)

	_, err := store.Get(cursorID)
	if err != ErrCursorExpired {
		t.Errorf("expected ErrCursorExpired, got %v", err)
	}
}

func TestCursorStore_WithFilter(t *testing.T) {
	store := NewCursorStore(30 * time.Minute)

	cursorID := store.Create([]byte("state"), "ks", "tbl", "id = 123", 100)

	cursor, err := store.Get(cursorID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cursor.Filter != "id = 123" {
		t.Errorf("expected filter 'id = 123', got %s", cursor.Filter)
	}
}
