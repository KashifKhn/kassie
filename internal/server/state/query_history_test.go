package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testQueryStore(t *testing.T) *QueryStore {
	t.Helper()
	return NewQueryStore(filepath.Join(t.TempDir(), "queries.json"))
}

func TestQueryStore_RecordHistory(t *testing.T) {
	store := testQueryStore(t)

	store.Record("p1", "SELECT 1")
	store.Record("p1", "SELECT 2")
	store.Record("p2", "SELECT x FROM other")

	got := store.History("p1", 10)
	if len(got) != 2 {
		t.Fatalf("history len = %d, want 2 (per-profile isolation)", len(got))
	}
	if got[0].CQL != "SELECT 2" {
		t.Errorf("most recent first: got %q, want SELECT 2", got[0].CQL)
	}
}

func TestQueryStore_HistoryLimitAndRing(t *testing.T) {
	store := testQueryStore(t)

	for i := range MaxHistoryEntries + 20 {
		store.Record("p", string(rune('a'+i%26))+"-"+string(rune(i)))
	}

	all := store.History("p", 0)
	if len(all) != MaxHistoryEntries {
		t.Fatalf("ring buffer size = %d, want %d", len(all), MaxHistoryEntries)
	}

	capped := store.History("p", 5)
	if len(capped) != 5 {
		t.Fatalf("limited history = %d, want 5", len(capped))
	}
}

func TestQueryStore_DedupsConsecutiveDuplicates(t *testing.T) {
	store := testQueryStore(t)

	store.Record("p", "SELECT 1")
	store.Record("p", "SELECT 1")
	store.Record("p", "SELECT 2")
	store.Record("p", "SELECT 1")

	got := store.History("p", 0)
	if len(got) != 3 {
		t.Fatalf("history = %d entries, want 3 (consecutive dup collapsed, repeat after gap kept)", len(got))
	}
	if got[0].CQL != "SELECT 1" || got[1].CQL != "SELECT 2" || got[2].CQL != "SELECT 1" {
		t.Errorf("unexpected order: %v", got)
	}
}

func TestQueryStore_ClearHistory(t *testing.T) {
	store := testQueryStore(t)

	store.Record("p", "SELECT 1")
	if n := store.ClearHistory("p"); n != 1 {
		t.Fatalf("cleared = %d, want 1", n)
	}
	if got := store.History("p", 0); len(got) != 0 {
		t.Fatalf("history not empty after clear: %v", got)
	}
	if n := store.ClearHistory("unknown"); n != 0 {
		t.Errorf("clearing unknown profile = %d, want 0", n)
	}
}

func TestQueryStore_SaveListDelete(t *testing.T) {
	store := testQueryStore(t)

	if err := store.SaveQuery("p", "top-users", "SELECT * FROM users LIMIT 10"); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := store.SaveQuery("p", "active-sessions", "SELECT count(*) FROM sessions"); err != nil {
		t.Fatalf("save: %v", err)
	}

	saved := store.SavedQueries("p")
	if len(saved) != 2 {
		t.Fatalf("saved = %d, want 2", len(saved))
	}
	if saved[0].Name > saved[1].Name {
		t.Errorf("saved queries not sorted: %q before %q", saved[0].Name, saved[1].Name)
	}

	if err := store.SaveQuery("p", "top-users", "SELECT * FROM users LIMIT 50"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	saved = store.SavedQueries("p")
	if len(saved) != 2 {
		t.Fatalf("upsert duplicated entry: %d", len(saved))
	}
	for _, sq := range saved {
		if sq.Name == "top-users" && sq.CQL != "SELECT * FROM users LIMIT 50" {
			t.Errorf("upsert did not update cql: %q", sq.CQL)
		}
	}

	if !store.DeleteSavedQuery("p", "top-users") {
		t.Fatal("delete returned false for existing query")
	}
	if store.DeleteSavedQuery("p", "top-users") {
		t.Error("delete returned true for missing query")
	}
	if got := store.SavedQueries("p"); len(got) != 1 {
		t.Fatalf("after delete saved = %d, want 1", len(got))
	}
}

func TestQueryStore_SaveQueryValidation(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{"ok-name_1.", nil},
		{"", ErrQueryNameInvalid},
		{strings.Repeat("x", 101), ErrQueryNameInvalid},
		{"bad;name", ErrQueryNameInvalid},
		{"bad/name", ErrQueryNameInvalid},
	}
	for _, tt := range tests {
		store := testQueryStore(t)
		err := store.SaveQuery("p", tt.name, "SELECT 1")
		if tt.wantErr == nil && err != nil {
			t.Errorf("name %q: unexpected error %v", tt.name, err)
		}
		if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
			t.Errorf("name %q: err = %v, want %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestQueryStore_PersistLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "queries.json")

	store1 := NewQueryStore(path)
	store1.Record("prod", "SELECT a FROM t1")
	if err := store1.SaveQuery("prod", "my-query", "SELECT b FROM t2"); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := store1.Persist(); err != nil {
		t.Fatalf("persist: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	store2 := NewQueryStore(path)
	if got := store2.History("prod", 10); len(got) != 1 || got[0].CQL != "SELECT a FROM t1" {
		t.Errorf("loaded history = %v", got)
	}
	saved := store2.SavedQueries("prod")
	if len(saved) != 1 || saved[0].Name != "my-query" || saved[0].CQL != "SELECT b FROM t2" {
		t.Errorf("loaded saved queries = %+v", saved)
	}
}

func TestQueryStore_PersistSkipsWhenClean(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queries.json")

	store := NewQueryStore(path)
	store.Record("p", "SELECT 1")
	if err := store.Persist(); err != nil {
		t.Fatalf("first persist: %v", err)
	}
	info1, _ := os.Stat(path)

	if err := store.Persist(); err != nil {
		t.Fatalf("second persist: %v", err)
	}
	info2, _ := os.Stat(path)

	if !info1.ModTime().Equal(info2.ModTime()) {
		t.Error("second persist rewrote unchanged file")
	}
}

func TestQueryStoreLatencyAggregation(t *testing.T) {
	store := testQueryStore(t)

	store.RecordLatency("p", "SELECT a", 100)
	store.RecordLatency("p", "SELECT a", 300)
	store.RecordLatency("p", "SELECT a", 200)

	slow := store.SlowQueries("p", 10)
	if len(slow) != 0 {
		t.Fatalf("fast query in slow list: %+v", slow)
	}

	hist := store.History("p", 10)
	if len(hist) != 1 {
		t.Fatalf("consecutive duplicate runs collapse: got %d", len(hist))
	}
	if hist[0].LatencyMs != 200 {
		t.Errorf("latest latency kept: %+v", hist[0])
	}
}

func TestQueryStoreSlowQueries(t *testing.T) {
	store := testQueryStore(t)

	store.RecordLatency("p", "SELECT fast", 50)
	store.RecordLatency("p", "SELECT slow1", 600)
	store.RecordLatency("p", "SELECT slow1", 900)
	store.RecordLatency("p", "SELECT slow1", 300)
	store.RecordLatency("p", "SELECT slow2", 2000)
	store.RecordLatency("other", "SELECT other", 5000)

	slow := store.SlowQueries("p", 10)
	if len(slow) != 2 {
		t.Fatalf("slow = %d, want 2 (fast excluded, per-profile)", len(slow))
	}

	if slow[0].CQL != "SELECT slow2" {
		t.Errorf("sorted by max latency: first = %+v", slow[0])
	}
	if slow[1].CQL != "SELECT slow1" {
		t.Errorf("second = %+v", slow[1])
	}

	s1 := slow[1]
	if s1.LastMs != 300 || s1.MaxMs != 900 || s1.ExecCount != 3 {
		t.Errorf("slow1 stats = %+v", s1)
	}
	if s1.AvgMs != 600 {
		t.Errorf("avg from recent samples = %d, want 600", s1.AvgMs)
	}
}

func TestQueryStoreSlowQueriesLimit(t *testing.T) {
	store := testQueryStore(t)

	for i := range 10 {
		store.RecordLatency("p", fmt.Sprintf("SELECT q%d", i), int64(600+i))
	}

	slow := store.SlowQueries("p", 3)
	if len(slow) != 3 {
		t.Fatalf("limit = %d, want 3", len(slow))
	}
	if slow[0].CQL != "SELECT q9" {
		t.Errorf("highest last latency first: %+v", slow[0])
	}
}

func TestQueryStoreSlowStatsPersist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queries.json")

	store1 := NewQueryStore(path)
	store1.RecordLatency("p", "SELECT slow", 800)
	store1.RecordLatency("p", "SELECT slow", 400)
	if err := store1.Persist(); err != nil {
		t.Fatalf("persist: %v", err)
	}

	store2 := NewQueryStore(path)
	slow := store2.SlowQueries("p", 10)
	if len(slow) != 1 {
		t.Fatalf("slow stats lost after reload: %+v", slow)
	}
	if slow[0].ExecCount != 2 || slow[0].MaxMs != 800 || slow[0].LastMs != 400 {
		t.Errorf("reloaded stats = %+v", slow[0])
	}
}
