package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"
)

const MaxHistoryEntries = 100

var ErrQueryNameInvalid = errors.New("query name must be 1-100 characters: letters, digits, spaces, dash, underscore, dot")

type HistoryEntry struct {
	CQL        string `json:"cql"`
	ExecutedAt int64  `json:"executed_at"`
}

type SavedQuery struct {
	Name      string `json:"name"`
	CQL       string `json:"cql"`
	CreatedAt int64  `json:"created_at"`
}

type queryFile struct {
	Version  int                     `json:"version"`
	Profiles map[string]*profileData `json:"profiles"`
}

type profileData struct {
	History []HistoryEntry `json:"history"`
	Saved   []SavedQuery   `json:"saved"`
}

type QueryStore struct {
	mu       sync.RWMutex
	profiles map[string]*profileData
	path     string
	dirty    bool
}

func NewQueryStore(path string) *QueryStore {
	store := &QueryStore{
		profiles: make(map[string]*profileData),
		path:     path,
	}
	_ = store.Load()
	return store
}

func DefaultQueryStorePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "kassie", "queries.json")
}

func (q *QueryStore) profile(name string) *profileData {
	pd, ok := q.profiles[name]
	if !ok {
		pd = &profileData{}
		q.profiles[name] = pd
	}
	return pd
}

func (q *QueryStore) Record(profile, cql string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if cql == "" {
		return
	}

	pd := q.profile(profile)
	entry := HistoryEntry{CQL: cql, ExecutedAt: time.Now().Unix()}

	if len(pd.History) > 0 && pd.History[len(pd.History)-1].CQL == cql {
		pd.History[len(pd.History)-1] = entry
	} else {
		pd.History = append(pd.History, entry)
	}

	if len(pd.History) > MaxHistoryEntries {
		pd.History = pd.History[len(pd.History)-MaxHistoryEntries:]
	}
	q.dirty = true
}

func (q *QueryStore) History(profile string, limit int) []HistoryEntry {
	q.mu.RLock()
	defer q.mu.RUnlock()

	pd, ok := q.profiles[profile]
	if !ok || len(pd.History) == 0 {
		return nil
	}

	result := make([]HistoryEntry, 0, len(pd.History))
	for i := len(pd.History) - 1; i >= 0; i-- {
		result = append(result, pd.History[i])
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func (q *QueryStore) ClearHistory(profile string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	pd, ok := q.profiles[profile]
	if !ok {
		return 0
	}
	n := len(pd.History)
	pd.History = nil
	if n > 0 {
		q.dirty = true
	}
	return n
}

var queryNameRegex = regexp.MustCompile(`^[a-zA-Z0-9 _.\-]{1,100}$`)

func (q *QueryStore) SaveQuery(profile, name, cql string) error {
	if !queryNameRegex.MatchString(name) {
		return ErrQueryNameInvalid
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	pd := q.profile(profile)
	for i := range pd.Saved {
		if pd.Saved[i].Name == name {
			if pd.Saved[i].CQL == cql {
				return nil
			}
			pd.Saved[i].CQL = cql
			q.dirty = true
			return nil
		}
	}

	pd.Saved = append(pd.Saved, SavedQuery{
		Name:      name,
		CQL:       cql,
		CreatedAt: time.Now().Unix(),
	})
	sort.Slice(pd.Saved, func(i, j int) bool { return pd.Saved[i].Name < pd.Saved[j].Name })
	q.dirty = true
	return nil
}

func (q *QueryStore) DeleteSavedQuery(profile, name string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	pd, ok := q.profiles[profile]
	if !ok {
		return false
	}

	for i, sq := range pd.Saved {
		if sq.Name == name {
			pd.Saved = append(pd.Saved[:i], pd.Saved[i+1:]...)
			q.dirty = true
			return true
		}
	}
	return false
}

func (q *QueryStore) SavedQueries(profile string) []SavedQuery {
	q.mu.RLock()
	defer q.mu.RUnlock()

	pd, ok := q.profiles[profile]
	if !ok {
		return nil
	}
	out := make([]SavedQuery, len(pd.Saved))
	copy(out, pd.Saved)
	return out
}

func (q *QueryStore) Load() error {
	data, err := os.ReadFile(q.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read queries file: %w", err)
	}

	var file queryFile
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("failed to parse queries file: %w", err)
	}
	if file.Profiles == nil {
		file.Profiles = make(map[string]*profileData)
	}

	q.mu.Lock()
	q.profiles = file.Profiles
	q.mu.Unlock()
	return nil
}

func (q *QueryStore) Persist() error {
	q.mu.Lock()
	if !q.dirty {
		q.mu.Unlock()
		return nil
	}

	file := queryFile{Version: 1, Profiles: q.profiles}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		q.mu.Unlock()
		return fmt.Errorf("failed to encode queries file: %w", err)
	}
	q.dirty = false
	q.mu.Unlock()

	dir := filepath.Dir(q.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	tmp := q.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("failed to write queries file: %w", err)
	}
	if err := os.Rename(tmp, q.path); err != nil {
		return fmt.Errorf("failed to replace queries file: %w", err)
	}
	return nil
}
