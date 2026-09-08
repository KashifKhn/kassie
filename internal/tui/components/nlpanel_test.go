package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func nlKeyMsg(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "ctrl+p":
		return tea.KeyMsg{Type: tea.KeyCtrlP}
	case "ctrl+m":
		return tea.KeyMsg{Type: tea.KeyCtrlE}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestNlPanelActivateDeactivate(t *testing.T) {
	panel := NewNlPanel()
	if panel.IsActive() {
		t.Fatal("fresh panel should be inactive")
	}

	panel, _ = panel.Activate("ks", "tbl")
	if !panel.IsActive() {
		t.Fatal("activated panel should be active")
	}

	panel = panel.Deactivate()
	if panel.IsActive() {
		t.Fatal("deactivated panel should be inactive")
	}
}

func TestNlPanelModelEditing(t *testing.T) {
	panel, _ := NewNlPanel().Activate("ks", "")
	if panel.Model() != "" {
		t.Fatalf("default model should be empty, got %q", panel.Model())
	}

	panel, _ = panel.Update(nlKeyMsg("ctrl+e"), nil)
	panel, _ = panel.Update(nlKeyMsg("gpt-x"), nil)
	panel, _ = panel.Update(nlKeyMsg("enter"), nil)
	if panel.Model() != "gpt-x" {
		t.Fatalf("model should be saved, got %q", panel.Model())
	}

	panel, _ = panel.Update(nlKeyMsg("ctrl+e"), nil)
	panel, _ = panel.Update(nlKeyMsg("backspace"), nil)
	panel, _ = panel.Update(nlKeyMsg("esc"), nil)
	if panel.Model() != "gpt-x" {
		t.Fatalf("esc should discard edits, got %q", panel.Model())
	}
}

func TestNlPanelModelEditingIsolated(t *testing.T) {
	panel, _ := NewNlPanel().Activate("ks", "")

	panel, _ = panel.Update(nlKeyMsg("ctrl+e"), nil)
	panel, _ = panel.Update(nlKeyMsg("h"), nil)
	panel, _ = panel.Update(nlKeyMsg("i"), nil)
	panel, _ = panel.Update(nlKeyMsg("enter"), nil)
	if panel.Model() != "hi" {
		t.Fatalf("model should be hi, got %q", panel.Model())
	}
	if string(panel.input) != "" {
		t.Fatalf("question input should be untouched during model editing, got %q", string(panel.input))
	}
}

func TestNlPanelProviderCycle(t *testing.T) {
	panel, _ := NewNlPanel().Activate("ks", "")
	if panel.Provider() != "auto" {
		t.Fatalf("default provider should be auto, got %s", panel.Provider())
	}

	panel, _ = panel.Update(nlKeyMsg("ctrl+p"), nil)
	if panel.Provider() != "opencode" {
		t.Fatalf("first cycle should be opencode, got %s", panel.Provider())
	}

	for i := 0; i < 4; i++ {
		panel, _ = panel.Update(nlKeyMsg("ctrl+p"), nil)
	}
	if panel.Provider() != "auto" {
		t.Fatalf("cycle should wrap to auto, got %s", panel.Provider())
	}
}

func TestNlPanelTypingAndSubmit(t *testing.T) {
	panel, _ := NewNlPanel().Activate("ks", "")

	panel, _ = panel.Update(nlKeyMsg("h"), nil)
	panel, _ = panel.Update(nlKeyMsg("i"), nil)

	if got := string(panel.input); got != "hi" {
		t.Fatalf("input = %q, want hi", got)
	}

	panel, _ = panel.Update(nlKeyMsg("backspace"), nil)
	if got := string(panel.input); got != "h" {
		t.Fatalf("after backspace input = %q, want h", got)
	}

	panel, _ = panel.Update(nlKeyMsg("i"), nil)
	newPanel, cmd := panel.Update(nlKeyMsg("enter"), nil)
	if !newPanel.loading {
		t.Fatal("submitting should set loading")
	}
	if cmd == nil {
		t.Fatal("submitting should return a command")
	}
}

func TestNlPanelEmptySubmitIgnored(t *testing.T) {
	panel, _ := NewNlPanel().Activate("ks", "")
	newPanel, cmd := panel.Update(nlKeyMsg("enter"), nil)
	if newPanel.loading || cmd != nil {
		t.Fatal("empty submit should be ignored")
	}
}

func TestNlPanelTranslatedAccept(t *testing.T) {
	panel, _ := NewNlPanel().Activate("ks", "")
	panel.translated = &NlTranslatedMsg{CQL: "SELECT 1", Provider: "mock", Question: "q", Keyspace: "ks"}

	newPanel, cmd := panel.Update(nlKeyMsg("enter"), nil)
	msg := cmd()
	translated, ok := msg.(NlTranslatedMsg)
	if !ok || translated.CQL != "SELECT 1" {
		t.Fatalf("expected NlTranslatedMsg, got %T", msg)
	}
	if newPanel.IsActive() {
		t.Fatal("panel should deactivate after accept")
	}
}

func TestNlPanelTranslatedDiscard(t *testing.T) {
	panel, _ := NewNlPanel().Activate("ks", "")
	panel.translated = &NlTranslatedMsg{CQL: "SELECT 1", Provider: "mock", Question: "q", Keyspace: "ks"}

	newPanel, cmd := panel.Update(nlKeyMsg("esc"), nil)
	msg := cmd()
	if _, ok := msg.(NlCanceledMsg); !ok {
		t.Fatalf("expected NlCanceledMsg, got %T", msg)
	}
	if newPanel.IsActive() {
		t.Fatal("panel should deactivate after discard")
	}
}
