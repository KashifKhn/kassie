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
	theme         styles.Theme
	input         textinput.Model
	active        bool
	validationErr string
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
	f.validationErr = ""
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
			if err := validateFilter(value); err != "" {
				f.validationErr = err
				return f, nil
			}
			f = f.Deactivate()
			return f, func() tea.Msg { return FilterAppliedMsg{Where: value} }
		case "esc":
			f = f.Deactivate()
			return f, func() tea.Msg { return FilterCanceledMsg{} }
		default:
			f.validationErr = ""
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

	if f.validationErr != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		bar += "\n" + errorStyle.Render("✗ "+f.validationErr)
	}

	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	content := lipgloss.NewStyle().Width(innerWidth).Render(bar)

	borderColor := lipgloss.Color("238")
	if f.validationErr != "" {
		borderColor = lipgloss.Color("196")
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Render(content)
}

func validateFilter(where string) string {
	where = strings.TrimSpace(where)

	if where == "" {
		return ""
	}

	upper := strings.ToUpper(where)

	dangerousKeywords := []string{"DROP", "DELETE", "TRUNCATE", "ALTER", "CREATE", "INSERT", "UPDATE"}
	for _, kw := range dangerousKeywords {
		if containsWholeWord(upper, kw) {
			return "Dangerous keyword '" + kw + "' not allowed in filters"
		}
	}

	singleQuotes := strings.Count(where, "'")
	if singleQuotes%2 != 0 {
		return "Unbalanced single quotes"
	}

	doubleQuotes := strings.Count(where, "\"")
	if doubleQuotes%2 != 0 {
		return "Unbalanced double quotes"
	}

	openParens := strings.Count(where, "(")
	closeParens := strings.Count(where, ")")
	if openParens != closeParens {
		return "Unbalanced parentheses"
	}

	validOperators := []string{"=", ">", "<", ">=", "<=", "!=", " IN ", " CONTAINS ", " LIKE ", " AND ", " OR "}
	hasOperator := false
	for _, op := range validOperators {
		if strings.Contains(upper, strings.ToUpper(op)) {
			hasOperator = true
			break
		}
	}

	if !hasOperator {
		return "WHERE clause must contain a valid operator (=, >, <, IN, etc.)"
	}

	return ""
}

func containsWholeWord(s, word string) bool {
	idx := strings.Index(s, word)
	if idx == -1 {
		return false
	}

	before := idx == 0 || !isAlphaNum(rune(s[idx-1]))
	after := idx+len(word) >= len(s) || !isAlphaNum(rune(s[idx+len(word)]))

	return before && after
}

func isAlphaNum(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_'
}
