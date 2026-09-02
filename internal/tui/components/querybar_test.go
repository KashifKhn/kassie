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
