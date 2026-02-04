package state

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCursorNotFound = errors.New("cursor not found")
	ErrCursorExpired  = errors.New("cursor expired")
)

type Cursor struct {
	ID        string
	PageState []byte
	Keyspace  string
	Table     string
	Filter    string
	PageSize  int
	CreatedAt time.Time
	LastUsed  time.Time
}

type CursorStore struct {
	cursors map[string]*Cursor
	mu      sync.RWMutex
	ttl     time.Duration
}

func NewCursorStore(ttl time.Duration) *CursorStore {
	store := &CursorStore{
		cursors: make(map[string]*Cursor),
		ttl:     ttl,
	}
	go store.cleanup()
	return store
}

func (cs *CursorStore) Create(pageState []byte, keyspace, table, filter string, pageSize int) string {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	id := uuid.New().String()
	cursor := &Cursor{
		ID:        id,
		PageState: pageState,
		Keyspace:  keyspace,
		Table:     table,
		Filter:    filter,
		PageSize:  pageSize,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	cs.cursors[id] = cursor
	return id
}

func (cs *CursorStore) Get(id string) (*Cursor, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	cursor, exists := cs.cursors[id]
	if !exists {
		return nil, ErrCursorNotFound
	}

	if time.Since(cursor.LastUsed) > cs.ttl {
		return nil, ErrCursorExpired
	}

	cursor.LastUsed = time.Now()
	return cursor, nil
}

func (cs *CursorStore) Delete(id string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.cursors, id)
}

func (cs *CursorStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cs.mu.Lock()
		now := time.Now()
		for id, cursor := range cs.cursors {
			if now.Sub(cursor.LastUsed) > cs.ttl {
				delete(cs.cursors, id)
			}
		}
		cs.mu.Unlock()
	}
}

func (cs *CursorStore) Count() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.cursors)
}
