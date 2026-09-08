package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const MaxNLExamples = 50

type NLExample struct {
	ID        string `json:"id"`
	Question  string `json:"question"`
	CQL       string `json:"cql"`
	Keyspace  string `json:"keyspace"`
	CreatedAt int64  `json:"created_at"`
}

type nlExampleFile struct {
	Version  int                   `json:"version"`
	Profiles map[string][]NLExample `json:"profiles"`
}

type NLExampleStore struct {
	mu       sync.RWMutex
	profiles map[string][]NLExample
	path     string
	dirty    bool
}

func NewNLExampleStore(path string) *NLExampleStore {
	store := &NLExampleStore{
		profiles: make(map[string][]NLExample),
		path:     path,
	}
	_ = store.Load()
	return store
}

func DefaultNLExampleStorePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "kassie", "nl-examples.json")
}

func (s *NLExampleStore) Add(profile, question, cql, keyspace string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if question == "" || cql == "" {
		return ""
	}

	id := fmt.Sprintf("nl-%d", time.Now().UnixNano())
	example := NLExample{
		ID:        id,
		Question:  question,
		CQL:       cql,
		Keyspace:  keyspace,
		CreatedAt: time.Now().Unix(),
	}

	s.profiles[profile] = append(s.profiles[profile], example)
	if len(s.profiles[profile]) > MaxNLExamples {
		s.profiles[profile] = s.profiles[profile][len(s.profiles[profile])-MaxNLExamples:]
	}
	s.dirty = true
	return id
}

func (s *NLExampleStore) List(profile, keyspace string, limit int) []NLExample {
	s.mu.RLock()
	defer s.mu.RUnlock()

	examples, ok := s.profiles[profile]
	if !ok || len(examples) == 0 {
		return nil
	}

	result := make([]NLExample, 0, len(examples))
	for i := len(examples) - 1; i >= 0; i-- {
		if keyspace != "" && examples[i].Keyspace != "" && examples[i].Keyspace != keyspace {
			continue
		}
		result = append(result, examples[i])
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func (s *NLExampleStore) Delete(profile, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	examples, ok := s.profiles[profile]
	if !ok {
		return false
	}

	for i, e := range examples {
		if e.ID == id {
			s.profiles[profile] = append(examples[:i], examples[i+1:]...)
			s.dirty = true
			return true
		}
	}
	return false
}

func (s *NLExampleStore) Load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read nl examples file: %w", err)
	}

	var file nlExampleFile
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("failed to parse nl examples file: %w", err)
	}
	if file.Profiles == nil {
		file.Profiles = make(map[string][]NLExample)
	}

	s.mu.Lock()
	s.profiles = file.Profiles
	s.mu.Unlock()
	return nil
}

func (s *NLExampleStore) Persist() error {
	s.mu.Lock()
	if !s.dirty {
		s.mu.Unlock()
		return nil
	}

	file := nlExampleFile{Version: 1, Profiles: s.profiles}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to encode nl examples file: %w", err)
	}
	s.dirty = false
	s.mu.Unlock()

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("failed to write nl examples file: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("failed to replace nl examples file: %w", err)
	}
	return nil
}
