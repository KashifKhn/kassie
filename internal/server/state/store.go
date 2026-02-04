package state

import (
	"errors"
	"sync"
	"time"

	"github.com/KashifKhn/kassie/internal/server/db"
	"github.com/KashifKhn/kassie/internal/shared/config"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

type Session struct {
	ID         string
	Profile    *config.Profile
	Connection *db.Session
	CreatedAt  time.Time
	LastAccess time.Time
	Cursors    *CursorStore
}

type Store struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	ttl      time.Duration
}

func NewStore(ttl time.Duration) *Store {
	store := &Store{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}
	go store.cleanup()
	return store
}

func (s *Store) Create(id string, profile *config.Profile, conn *db.Session) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &Session{
		ID:         id,
		Profile:    profile,
		Connection: conn,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
		Cursors:    NewCursorStore(30 * time.Minute),
	}

	s.sessions[id] = session
	return session
}

func (s *Store) Get(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[id]
	if !exists {
		return nil, ErrSessionNotFound
	}

	if time.Since(session.LastAccess) > s.ttl {
		return nil, ErrSessionExpired
	}

	session.LastAccess = time.Now()
	return session, nil
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[id]; exists {
		if session.Connection != nil {
			session.Connection.Close()
		}
		delete(s.sessions, id)
	}
}

func (s *Store) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, session := range s.sessions {
			if now.Sub(session.LastAccess) > s.ttl {
				if session.Connection != nil {
					session.Connection.Close()
				}
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}
