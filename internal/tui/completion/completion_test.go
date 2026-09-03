package completion

import (
	"strings"
	"testing"
)

func testSources() Sources {
	return Sources{
		DefaultKeyspace: "app_data",
		DefaultTable:    "users",
		Keyspaces:       []string{"app_data", "audit", "system_auth"},
		TablesFor: func(ks string) []string {
			if ks == "app_data" {
				return []string{"users", "sessions", "user_events"}
			}
			if ks == "" {
				return []string{"current_table"}
			}
			return nil
		},
		ColumnsFor: func(ks, tbl string) []Column {
			if ks == "app_data" && tbl == "users" {
				return []Column{
					{Name: "id", CqlType: "uuid"},
					{Name: "email", CqlType: "text"},
					{Name: "created_at", CqlType: "timestamp"},
					{Name: "region", CqlType: "text"},
				}
			}
			return nil
		},
	}
}

func labels(s []Suggestion) string {
	parts := make([]string, len(s))
	for i, x := range s {
		parts[i] = x.Label
	}
	return strings.Join(parts, ",")
}

func TestCompleteEmptyText(t *testing.T) {
	got := Complete("", testSources())
	if !strings.Contains(labels(got), "SELECT") {
		t.Errorf("empty text should suggest SELECT keyword, got %q", labels(got))
	}
}

func TestCompleteStartOfStatement(t *testing.T) {
	got := Complete("SE", testSources())
	if len(got) == 0 || got[0].Label != "SELECT" {
		t.Errorf("SE should complete to SELECT, got %q", labels(got))
	}
}

func TestCompleteAfterFromSuggestsKeyspaces(t *testing.T) {
	got := Complete("SELECT * FROM ap", testSources())
	if !strings.Contains(labels(got), "app_data") {
		t.Errorf("FROM ap should suggest app_data keyspace, got %q", labels(got))
	}
}

func TestCompleteAfterFromSpaceSuggestsKeyspaces(t *testing.T) {
	got := Complete("SELECT * FROM ", testSources())
	all := labels(got)
	if !strings.Contains(all, "app_data") || !strings.Contains(all, "audit") {
		t.Errorf("FROM + space should suggest keyspaces, got %q", all)
	}
}

func TestCompleteKeyspaceDotSuggestsTables(t *testing.T) {
	got := Complete("SELECT * FROM app_data.u", testSources())
	gotLabels := labels(got)
	if !strings.Contains(gotLabels, "app_data.users") {
		t.Errorf("app_data.u should suggest app_data.users, got %q", gotLabels)
	}
	if strings.Contains(gotLabels, "app_data.sessions") {
		t.Errorf("u prefix should not match sessions: %q", gotLabels)
	}
}

func TestCompleteKeyspaceDotAllTables(t *testing.T) {
	got := Complete("SELECT * FROM app_data.", testSources())
	all := labels(got)
	if !strings.Contains(all, "app_data.users") || !strings.Contains(all, "app_data.sessions") {
		t.Errorf("app_data. should list all tables, got %q", all)
	}
}

func TestCompleteWhereColumns(t *testing.T) {
	got := Complete("SELECT * FROM app_data.users WHERE re", testSources())
	if !strings.Contains(labels(got), "region") {
		t.Errorf("WHERE re should suggest region column, got %q", labels(got))
	}
}

func TestCompleteWhereColumnsWithDetail(t *testing.T) {
	got := Complete("SELECT * FROM app_data.users WHERE e", testSources())
	for _, s := range got {
		if s.Label == "email" {
			if s.Detail != "text" || s.Kind != KindColumn {
				t.Errorf("email suggestion = %+v, want kind column with text type", s)
			}
			return
		}
	}
	t.Errorf("email not suggested: %q", labels(got))
}

func TestCompleteSelectColumns(t *testing.T) {
	got := Complete("SELECT em", testSources())
	if !strings.Contains(labels(got), "email") {
		t.Errorf("SELECT em should suggest email column, got %q", labels(got))
	}
}

func TestCompleteWhereKeywordsAsFallback(t *testing.T) {
	got := Complete("SELECT * FROM app_data.users WHERE region = 'eu' AN", testSources())
	if !strings.Contains(labels(got), "AND") {
		t.Errorf("AN should suggest AND keyword, got %q", labels(got))
	}
}

func TestResolveTargetTable(t *testing.T) {
	tests := []struct {
		text   string
		wantKs string
		wantTb string
	}{
		{"SELECT * FROM app_data.users", "app_data", "users"},
		{"SELECT * FROM app_data.users WHERE id = 1", "app_data", "users"},
		{"select id from audit.log", "audit", "log"},
		{"SELECT * FROM users", "", "users"},
		{"SELECT 1", "", ""},
	}
	for _, tt := range tests {
		ks, tbl := resolveTargetTable(tt.text)
		if ks != tt.wantKs || tbl != tt.wantTb {
			t.Errorf("resolveTargetTable(%q) = (%q,%q), want (%q,%q)", tt.text, ks, tbl, tt.wantKs, tt.wantTb)
		}
	}
}

func TestSuggestionKindString(t *testing.T) {
	if KindKeyword.String() != "keyword" || KindKeyspace.String() != "keyspace" {
		t.Error("kind strings wrong")
	}
}
