package components

import (
	"strings"

	"github.com/KashifKhn/kassie/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type QuerySubmittedMsg struct {
	CQL string
}

type QueryCanceledMsg struct{}

type QueryBar struct {
	theme         styles.Theme
	input         textinput.Model
	active        bool
	validationErr string
}

func NewQueryBar(theme styles.Theme) QueryBar {
	input := textinput.New()
	input.Placeholder = "SELECT * FROM keyspace.table"
	input.Prompt = "» "
	input.CharLimit = 10000
	input.Width = 60
	return QueryBar{theme: theme, input: input}
}

func (q QueryBar) Activate() QueryBar {
	q.active = true
	q.validationErr = ""
	q.input.CursorEnd()
	q.input.Focus()
	return q
}

func (q QueryBar) Deactivate() QueryBar {
	q.active = false
	q.input.Blur()
	return q
}

func (q QueryBar) IsActive() bool {
	return q.active
}

func (q QueryBar) Value() string {
	return strings.TrimSpace(q.input.Value())
}

func (q QueryBar) Update(msg tea.Msg) (QueryBar, tea.Cmd) {
	if !q.active {
		return q, nil
	}

	var cmd tea.Cmd
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "enter":
			value := q.Value()
			if err := validateQueryCQL(value); err != "" {
				q.validationErr = err
				return q, nil
			}
			q = q.Deactivate()
			return q, func() tea.Msg { return QuerySubmittedMsg{CQL: value} }
		case "esc":
			q = q.Deactivate()
			return q, func() tea.Msg { return QueryCanceledMsg{} }
		default:
			q.validationErr = ""
		}
	}

	q.input, cmd = q.input.Update(msg)
	return q, cmd
}

func (q QueryBar) View(width int) string {
	if !q.active {
		return ""
	}

	inputWidth := width - 8
	if inputWidth < 10 {
		inputWidth = 10
	}
	q.input.Width = inputWidth

	borderColor := "51"
	if q.validationErr != "" {
		borderColor = "196"
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Width(width - 2)

	content := q.input.View()
	if q.validationErr != "" {
		content += "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render("✗ " + q.validationErr)
	}

	return box.Render(content)
}

func validateQueryCQL(cql string) string {
	if cql == "" {
		return "query is empty"
	}
	body := strings.TrimRight(cql, "; \t\n\r")
	if strings.Contains(body, ";") {
		return "multiple statements are not allowed"
	}

	upper := strings.ToUpper(body)
	if !strings.HasPrefix(upper, "SELECT") {
		return "only SELECT statements are allowed"
	}

	for _, kw := range []string{"INSERT", "UPDATE", "DELETE", "ALTER", "DROP", "CREATE", "TRUNCATE", "GRANT", "REVOKE", "BATCH", "USE", "BEGIN"} {
		if containsWholeWord(upper, kw) {
			return "keyword " + kw + " is not allowed; read-only SELECT only"
		}
	}

	return ""
}
