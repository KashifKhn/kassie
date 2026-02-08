package components

import (
	"fmt"

	"github.com/KashifKhn/kassie/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
	theme styles.Theme
}

func NewStatusBar(theme styles.Theme) StatusBar {
	return StatusBar{theme: theme}
}

func (s StatusBar) View(width int, profile, keyspace, table, status string) string {
	left := profile
	if keyspace != "" && table != "" {
		left = fmt.Sprintf("%s › %s › %s", profile, keyspace, table)
	} else if keyspace != "" {
		left = fmt.Sprintf("%s › %s", profile, keyspace)
	}
	if left == "" {
		left = "Not connected"
	}

	content := fmt.Sprintf("%s │ %s", left, status)
	return lipgloss.NewStyle().Width(width).Render(s.theme.Status.Render(content))
}
