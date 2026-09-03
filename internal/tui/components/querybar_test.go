package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/KashifKhn/kassie/internal/tui/styles"
)

func TestValidateQueryCQL(t *testing.T) {
	tests := []struct {
		name    string
		cql     string
		wantErr string
	}{
		{"valid select", "SELECT * FROM ks.tbl", ""},
		{"trailing semicolon ok", "SELECT id FROM ks.t;", ""},
		{"empty", "", "empty"},
		{"multi statement", "SELECT 1; DROP TABLE x", "multiple statements"},
		{"not select", "INSERT INTO t (a) VALUES (1)", "only SELECT"},
		{"delete", "DELETE FROM t", "only SELECT"},
		{"use keyword", "SELECT * FROM use", "not allowed"},
		{"identifier substring ok", "SELECT used_count, insertions FROM metrics_daily", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQueryCQL(tt.cql)
			if tt.wantErr == "" {
				if err != "" {
					t.Errorf("unexpected error: %q", err)
				}
				return
			}
			if !strings.Contains(err, tt.wantErr) {
				t.Errorf("err = %q, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestQueryBarLifecycle(t *testing.T) {
	bar := NewQueryBar(styles.DefaultTheme())
	if bar.IsActive() {
		t.Fatal("starts inactive")
	}

	bar = bar.Activate()
	if !bar.IsActive() {
		t.Fatal("active after Activate")
	}

	updated, _ := bar.Update(teabt("SELECT 1"))
	if !strings.Contains(updated.Value(), "SELECT 1") {
		t.Errorf("value = %q", updated.Value())
	}

	updated, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if updated.IsActive() {
		t.Error("deactivated after submit")
	}
	if cmd == nil {
		t.Fatal("submit emits msg")
	}
	msg := cmd()
	if _, ok := msg.(QuerySubmittedMsg); !ok {
		t.Errorf("msg = %T, want QuerySubmittedMsg", msg)
	}
}

func TestQueryBarCancel(t *testing.T) {
	bar := NewQueryBar(styles.DefaultTheme()).Activate()
	bar, cmd := bar.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if bar.IsActive() {
		t.Error("deactivated after esc")
	}
	if _, ok := cmd().(QueryCanceledMsg); !ok {
		t.Errorf("msg = %T, want QueryCanceledMsg", cmd())
	}
}

func TestQueryBarInvalidKeepsOpen(t *testing.T) {
	bar := NewQueryBar(styles.DefaultTheme()).Activate()
	bar, _ = bar.Update(teabt("DROP TABLE x"))
	bar, cmd := bar.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !bar.IsActive() {
		t.Error("invalid query must keep bar open")
	}
	if cmd != nil {
		t.Error("no msg on validation error")
	}
	if bar.validationErr == "" {
		t.Error("validation error surfaced")
	}
}

func teabt(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestQueryBarSuggestionsAndTab(t *testing.T) {
	bar := NewQueryBar(styles.DefaultTheme())
	bar.SetSources([]string{"app_data"}, func(ks string) []string {
		if ks == "app_data" {
			return []string{"users"}
		}
		return nil
	}, nil)
	bar = bar.Activate()

	bar, _ = bar.Update(teabt("SELECT * FROM app"))
	if len(bar.suggestions) == 0 {
		t.Fatal("expected keyspace suggestion for 'app'")
	}
	if bar.suggestions[0].Label != "app_data" {
		t.Errorf("first suggestion = %q, want app_data", bar.suggestions[0].Label)
	}

	bar, _ = bar.Update(teaKey(tea.KeyTab))
	value := bar.input.Value()
	if !strings.HasSuffix(value, "app_data ") {
		t.Errorf("tab accept: value = %q, want suffix 'app_data '", value)
	}
}

func TestQueryBarUpDownCyclesSuggestions(t *testing.T) {
	bar := NewQueryBar(styles.DefaultTheme())
	bar = bar.Activate()
	bar, _ = bar.Update(teabt("A"))

	if len(bar.suggestions) < 2 {
		t.Fatalf("expected multiple keyword suggestions, got %d", len(bar.suggestions))
	}

	bar, _ = bar.Update(teaKey(tea.KeyDown))
	if bar.suggestSel != 1 {
		t.Errorf("after down: sel = %d, want 1", bar.suggestSel)
	}
	bar, _ = bar.Update(teaKey(tea.KeyUp))
	if bar.suggestSel != 0 {
		t.Errorf("after up: sel = %d, want 0", bar.suggestSel)
	}
}

func TestQueryBarViewShowsSuggestions(t *testing.T) {
	bar := NewQueryBar(styles.DefaultTheme())
	bar = bar.Activate()
	bar, _ = bar.Update(teabt("SELECT"))

	out := bar.View(80)
	if !strings.Contains(out, "SELECT") {
		t.Errorf("suggestion line missing: %q", out)
	}
	if !strings.Contains(out, "▸") {
		t.Errorf("selection marker missing: %q", out)
	}
	if !strings.Contains(out, "(keyword)") {
		t.Errorf("kind annotation missing: %q", out)
	}
}

func teaKey(key tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: key}
}
