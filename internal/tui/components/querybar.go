package components

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/tui/completion"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type QuerySubmittedMsg struct {
	CQL string
}

type QueryCanceledMsg struct{}

type QueryBarKeyspacesLoadedMsg struct {
	Keyspaces []string
}

func FetchQueryBarKeyspacesCmd(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := c.ListKeyspaces(ctx)
		if err != nil {
			return QueryBarKeyspacesLoadedMsg{}
		}

		names := make([]string, 0, len(resp))
		for _, ks := range resp {
			names = append(names, ks.Name)
		}
		return QueryBarKeyspacesLoadedMsg{Keyspaces: names}
	}
}

// FetchQueryBarTablesCmd prefetches table names for one keyspace so the
// bar's table lister can serve them synchronously.
func FetchQueryBarTablesCmd(c *client.Client, keyspace string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := c.ListTables(ctx, keyspace)
		if err != nil {
			return QueryBarTablesLoadedMsg{}
		}

		names := make([]string, 0, len(resp))
		for _, t := range resp {
			names = append(names, t.Name)
		}
		return QueryBarTablesLoadedMsg{Keyspace: keyspace, Tables: names}
	}
}

type QueryBarTablesLoadedMsg struct {
	Keyspace string
	Tables   []string
}

type TableLister func(keyspace string) []string
type ColumnLister func(keyspace, table string) []completion.Column

type QueryBar struct {
	theme         styles.Theme
	input         textinput.Model
	active        bool
	validationErr string

	suggestSel      int
	suggestions     []completion.Suggestion
	keyspaces       []string
	tablesCache     map[string][]string
	tableLister     TableLister
	columnLister    ColumnLister
	defaultKeyspace string
	defaultTable    string
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
	q.suggestions = nil
	q.suggestSel = 0
	q.input.CursorEnd()
	q.input.Focus()
	return q
}

func (q QueryBar) Deactivate() QueryBar {
	q.active = false
	q.suggestions = nil
	q.suggestSel = 0
	q.input.Blur()
	return q
}

func (q QueryBar) IsActive() bool {
	return q.active
}

func (q QueryBar) Value() string {
	return strings.TrimSpace(q.input.Value())
}

func (q *QueryBar) SetDefaultTable(keyspace, table string) {
	q.defaultKeyspace = keyspace
	q.defaultTable = table
}

func (q *QueryBar) SetSources(keyspaces []string, tables TableLister, columns ColumnLister) {
	q.keyspaces = keyspaces
	q.tableLister = tables
	q.columnLister = columns
}

func (q *QueryBar) SetColumnLister(columns ColumnLister) {
	q.columnLister = columns
}

func (q QueryBar) Update(msg tea.Msg) (QueryBar, tea.Cmd) {
	if !q.active {
		return q, nil
	}

	var cmd tea.Cmd
	switch m := msg.(type) {
	case QueryBarKeyspacesLoadedMsg:
		q.keyspaces = m.Keyspaces
		return q, nil
	case QueryBarTablesLoadedMsg:
		if q.tablesCache == nil {
			q.tablesCache = map[string][]string{}
		}
		q.tablesCache[m.Keyspace] = m.Tables
		return q, nil
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
		case "tab":
			if len(q.suggestions) > 0 {
				q = q.applySuggestion(q.suggestions[q.suggestSel])
				return q, nil
			}
			q = q.refreshSuggestions()
			return q, nil
		case "down":
			if len(q.suggestions) > 1 {
				q.suggestSel = (q.suggestSel + 1) % len(q.suggestions)
			}
			return q, nil
		case "up":
			if len(q.suggestions) > 1 {
				q.suggestSel = (q.suggestSel - 1 + len(q.suggestions)) % len(q.suggestions)
			}
			return q, nil
		case "ctrl+space":
			q = q.refreshSuggestions()
			return q, nil
		default:
			q.validationErr = ""
			q.input, cmd = q.input.Update(msg)
			q = q.refreshSuggestions()
			return q, cmd
		}
	}

	q.input, cmd = q.input.Update(msg)
	return q, cmd
}

func (q QueryBar) applySuggestion(s completion.Suggestion) QueryBar {
	text := q.input.Value()
	prefix, _ := completion.Analyze(text)

	// Replace the trailing partial token with the suggestion label.
	// For keyspace-dot suggestions the label is "ks.table" — drop "ks." part
	// that's already typed.
	label := s.Label
	if strings.Contains(label, ".") && strings.HasSuffix(text, ".") {
		label = strings.SplitN(label, ".", 2)[1]
	}

	if prefix == "" {
		q.input.SetValue(text + label + " ")
	} else {
		q.input.SetValue(text[:len(text)-len(prefix)] + label + " ")
	}
	q.input.CursorEnd()

	q.suggestions = nil
	q.suggestSel = 0
	q = q.refreshSuggestions()
	return q
}

func (q QueryBar) refreshSuggestions() QueryBar {
	sources := completion.Sources{
		Keyspaces:       q.keyspaces,
		DefaultKeyspace: q.defaultKeyspace,
		DefaultTable:    q.defaultTable,
	}
	if q.tablesCache != nil {
		sources.TablesFor = func(keyspace string) []string {
			return q.tablesCache[keyspace]
		}
	} else if q.tableLister != nil {
		sources.TablesFor = q.tableLister
	}
	if q.columnLister != nil {
		sources.ColumnsFor = q.columnLister
	}

	q.suggestions = completion.Complete(q.input.Value(), sources)
	q.suggestSel = 0
	return q
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

	if len(q.suggestions) > 0 {
		content += "\n" + q.renderSuggestions(width-8)
	}

	if q.validationErr != "" {
		content += "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render("✗ "+q.validationErr)
	}

	return box.Render(content)
}

func (q QueryBar) renderSuggestions(maxWidth int) string {
	kindColor := map[completion.SuggestionKind]string{
		completion.KindKeyword:  "240",
		completion.KindKeyspace: "51",
		completion.KindTable:    "82",
		completion.KindColumn:   "226",
	}

	const maxShown = 5
	shown := q.suggestions
	if len(shown) > maxShown {
		shown = shown[:maxShown]
	}

	parts := make([]string, 0, len(shown)+1)
	for i, s := range shown {
		label := s.Label
		detail := ""
		if s.Detail != "" {
			detail = " " + s.Detail
		}
		entry := label + detail + " (" + s.Kind.String() + ")"
		if len(entry) > maxWidth {
			entry = entry[:maxWidth-1] + "…"
		}

		if i == q.suggestSel {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Bold(true).
				Render("▸ "+entry))
		} else {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(kindColor[s.Kind]))
			parts = append(parts, style.Render("  "+entry))
		}
	}
	if len(q.suggestions) > maxShown {
		parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf("  … %d more (Tab cycle)", len(q.suggestions)-maxShown)))
	}

	return strings.Join(parts, "\n")
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
