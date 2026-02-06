package components

import (
	"strings"

	"github.com/KashifKhn/kassie/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FilterAppliedMsg struct {
	Where string
}

type FilterCanceledMsg struct{}

type FilterBar struct {
	theme  styles.Theme
	input  textinput.Model
	active bool
}

func NewFilterBar(theme styles.Theme) FilterBar {
	input := textinput.New()
	input.Placeholder = "WHERE clause"
	input.Prompt = "/ "
	input.CharLimit = 500
	input.Width = 40

	return FilterBar{
		theme: theme,
		input: input,
	}
}

func (f FilterBar) Activate(current string) FilterBar {
	f.active = true
	f.input.SetValue(current)
	f.input.CursorEnd()
	f.input.Width = maxInt(10, f.input.Width)
	f.input.Focus()
	return f
}

func (f FilterBar) Deactivate() FilterBar {
	f.active = false
	f.input.Blur()
	return f
}

func (f FilterBar) IsActive() bool {
	return f.active
}

func (f FilterBar) Value() string {
	return strings.TrimSpace(f.input.Value())
}

func (f FilterBar) Update(msg tea.Msg) (FilterBar, tea.Cmd) {
	if !f.active {
		return f, nil
	}

	var cmd tea.Cmd
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "enter":
			value := f.Value()
			f = f.Deactivate()
			return f, func() tea.Msg { return FilterAppliedMsg{Where: value} }
		case "esc":
			f = f.Deactivate()
			return f, func() tea.Msg { return FilterCanceledMsg{} }
		}
	}

	f.input, cmd = f.input.Update(msg)
	return f, cmd
}

func (f FilterBar) View(width int) string {
	if !f.active {
		return ""
	}

	if width > 4 {
		f.input.Width = width - 4
	}

	bar := f.input.View()
	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	content := lipgloss.NewStyle().Width(innerWidth).Render(bar)
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238")).Width(width).Render(content)
}
