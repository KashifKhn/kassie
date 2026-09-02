package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/KashifKhn/kassie/internal/tui/styles"
)

func TestQueryListNavigation(t *testing.T) {
	list := NewQueryList(styles.DefaultTheme())
	list.SetItems(QueryListHistory, []queryListItem{
		{cql: "SELECT 1"},
		{cql: "SELECT 2"},
		{cql: "SELECT 3"},
	})
	list.active = true

	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, nil)
	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, nil)
	if list.selected != 2 {
		t.Fatalf("selected = %d, want 2", list.selected)
	}

	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}, nil)
	if list.selected != 1 {
		t.Fatalf("selected = %d, want 1", list.selected)
	}

	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")}, nil)
	if list.selected != 2 {
		t.Fatalf("G selected = %d, want 2", list.selected)
	}

	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}, nil)
	if list.selected != 0 {
		t.Fatalf("g selected = %d, want 0", list.selected)
	}
}

func TestQueryListEnterPicks(t *testing.T) {
	list := NewQueryList(styles.DefaultTheme())
	list.SetItems(QueryListHistory, []queryListItem{{cql: "SELECT * FROM t"}})
	list.active = true

	updated, cmd := list.Update(tea.KeyMsg{Type: tea.KeyEnter}, nil)
	if updated.IsActive() {
		t.Error("deactivated after pick")
	}
	msg := cmd()
	picked, ok := msg.(QueryPickedMsg)
	if !ok {
		t.Fatalf("msg = %T, want QueryPickedMsg", msg)
	}
	if picked.CQL != "SELECT * FROM t" {
		t.Errorf("picked = %q", picked.CQL)
	}
}

func TestQueryListEscCloses(t *testing.T) {
	list, _ := NewQueryList(styles.DefaultTheme()).Activate(QueryListSaved)
	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyEsc}, nil)
	if list.IsActive() {
		t.Error("esc must close list")
	}
}

func TestQueryListDeleteWithoutClientSafe(t *testing.T) {
	list := NewQueryList(styles.DefaultTheme())
	list.SetItems(QueryListSaved, []queryListItem{{name: "q1", cql: "SELECT 1"}})
	list.active = true

	_, cmd := list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}, nil)
	if cmd == nil {
		t.Fatal("delete returns cmd")
	}
	msg := cmd()
	if _, ok := msg.(dataErrMsg); !ok {
		t.Errorf("nil client should yield error msg, got %T", msg)
	}
}

func TestQueryListViewRendersItemsAndSelection(t *testing.T) {
	list := NewQueryList(styles.DefaultTheme())
	list.SetItems(QueryListSaved, []queryListItem{
		{name: "top", cql: "SELECT a"},
		{name: "count", cql: "SELECT count(*)"},
	})
	list.active = true

	out := list.View(80)
	if !strings.Contains(out, "Saved Queries") {
		t.Errorf("title missing: %q", out)
	}
	if !strings.Contains(out, "top: SELECT a") {
		t.Errorf("item missing: %q", out)
	}
	if !strings.Contains(out, "▸") {
		t.Error("selection marker missing")
	}
	if !strings.Contains(out, "d delete (saved)") {
		t.Errorf("hint missing: %q", out)
	}
}

func TestQueryListEmptyState(t *testing.T) {
	list, _ := NewQueryList(styles.DefaultTheme()).Activate(QueryListHistory)
	out := list.View(80)
	if !strings.Contains(out, "empty") {
		t.Errorf("empty state missing: %q", out)
	}
}
