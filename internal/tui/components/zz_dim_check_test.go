package components

import (
	"strings"
	"testing"

	"github.com/KashifKhn/kassie/internal/tui/completion"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestSuggestionDimANSI(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(prev)

	q := QueryBar{
		suggestions: []completion.Suggestion{
			{Label: "email", Detail: "text", Kind: completion.KindColumn},
			{Label: "SELECT", Kind: completion.KindKeyword},
		},
		suggestSel: 0,
	}

	out := q.renderSuggestions(60)
	if !strings.Contains(out, "\x1b[38;5;240m") {
		t.Errorf("dim color (240) missing from output:\n%q", out)
	}

	lines := strings.Split(out, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[1], "\x1b[38;5;240m") {
		t.Errorf("second line missing dim suffix: %q", lines[1])
	}
}
