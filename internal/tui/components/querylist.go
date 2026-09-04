package components

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type QueryListMode int

const (
	QueryListHistory QueryListMode = iota
	QueryListSaved
	QueryListSlow
)

type QueryPickedMsg struct {
	CQL string
}

type QueryListClosedMsg struct{}

type queryListItem struct {
	name   string
	cql    string
	detail string
}

type QueryList struct {
	theme    styles.Theme
	active   bool
	mode     QueryListMode
	items    []queryListItem
	selected int
	err      string
}

func NewQueryList(theme styles.Theme) QueryList {
	return QueryList{theme: theme}
}

func (q QueryList) IsActive() bool {
	return q.active
}

func (q QueryList) Mode() QueryListMode {
	return q.mode
}

func (q QueryList) Activate(mode QueryListMode) (QueryList, tea.Cmd) {
	q.active = true
	q.mode = mode
	q.selected = 0
	q.err = ""
	q.items = nil
	return q, fetchQueryListCmd(nil, mode)
}

func (q QueryList) Deactivate() QueryList {
	q.active = false
	q.items = nil
	q.err = ""
	return q
}

func (q QueryList) Update(msg tea.Msg, c *client.Client) (QueryList, tea.Cmd) {
	if !q.active {
		return q, nil
	}

	switch m := msg.(type) {
	case queryListLoadedMsg:
		q.items = m.items
		q.err = ""
		return q, nil
	case dataErrMsg:
		q.err = m.Err.Error()
		return q, nil
	}

	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return q, nil
	}

	switch key.String() {
	case "j", "down":
		if q.selected < len(q.items)-1 {
			q.selected++
		}
	case "k", "up":
		if q.selected > 0 {
			q.selected--
		}
	case "g":
		q.selected = 0
	case "G":
		if len(q.items) > 0 {
			q.selected = len(q.items) - 1
		}
	case "d":
		if q.mode == QueryListSaved && q.selected < len(q.items) {
			item := q.items[q.selected]
			return q, deleteSavedQueryCmd(c, item.name)
		}
	case "ctrl+s":
		if q.mode != QueryListSlow {
			return q, nil
		}
		q = q.Deactivate()
		return q, func() tea.Msg { return QueryListClosedMsg{} }
	case "enter":
		if q.selected < len(q.items) {
			cql := q.items[q.selected].cql
			q = q.Deactivate()
			return q, func() tea.Msg { return QueryPickedMsg{CQL: cql} }
		}
	case "esc", "ctrl+y", "ctrl+p":
		q = q.Deactivate()
		return q, func() tea.Msg { return QueryListClosedMsg{} }
	}

	return q, nil
}

func (q QueryList) View(width int) string {
	if !q.active {
		return ""
	}

	title := "Query History"
	switch q.mode {
	case QueryListSaved:
		title = "Saved Queries"
	case QueryListSlow:
		title = "Slow Queries (>500ms)"
	}

	borderColor := "51"
	bodyLines := make([]string, 0, len(q.items)+1)
	bodyLines = append(bodyLines, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51")).Render(title))

	if q.err != "" {
		borderColor = "196"
		bodyLines = append(bodyLines, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗ "+q.err))
	}

	if len(q.items) == 0 && q.err == "" {
		bodyLines = append(bodyLines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("empty"))
	}

	maxWidth := width - 6
	if maxWidth < 20 {
		maxWidth = 20
	}

	for i, item := range q.items {
		label := item.cql
		if q.mode == QueryListSaved && item.name != "" {
			label = item.name + ": " + item.cql
		}
		if item.detail != "" {
			label = label + "  " + item.detail
		}
		if len(label) > maxWidth {
			label = label[:maxWidth-3] + "..."
		}

		if i == q.selected {
			bodyLines = append(bodyLines, lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Bold(true).
				Render("▸ "+label))
		} else {
			bodyLines = append(bodyLines, "  "+label)
		}
	}

	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).
		Render("j/k navigate • enter run • d delete (saved) • esc close")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Width(width - 2).
		Render(strings.Join(append(bodyLines, hint), "\n"))
}

type queryListLoadedMsg struct {
	items []queryListItem
}

var errNoClient = errors.New("no client connection")

func fetchQueryListCmd(c *client.Client, mode QueryListMode) tea.Cmd {
	return func() tea.Msg {
		if c == nil {
			return dataErrMsg{Err: errNoClient}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if mode == QueryListHistory {
			entries, err := c.ListQueryHistory(ctx, 50)
			if err != nil {
				return dataErrMsg{Err: err}
			}
			items := make([]queryListItem, 0, len(entries))
			for _, e := range entries {
				items = append(items, queryListItem{cql: e.Cql})
			}
			return queryListLoadedMsg{items: items}
		}

		if mode == QueryListSlow {
			slow, err := c.GetSlowQueries(ctx, 20)
			if err != nil {
				return dataErrMsg{Err: err}
			}
			items := make([]queryListItem, 0, len(slow))
			for _, sq := range slow {
				detail := fmt.Sprintf("last %dms  avg %dms  max %dms  ×%d",
					sq.LastLatencyMs, sq.AvgLatencyMs, sq.MaxLatencyMs, sq.ExecCount)
				items = append(items, queryListItem{cql: sq.Cql, detail: detail})
			}
			return queryListLoadedMsg{items: items}
		}

		saved, err := c.ListSavedQueries(ctx)
		if err != nil {
			return dataErrMsg{Err: err}
		}
		items := make([]queryListItem, 0, len(saved))
		for _, sq := range saved {
			items = append(items, queryListItem{name: sq.Name, cql: sq.Cql})
		}
		return queryListLoadedMsg{items: items}
	}
}

func deleteSavedQueryCmd(c *client.Client, name string) tea.Cmd {
	return func() tea.Msg {
		if c == nil {
			return dataErrMsg{Err: errNoClient}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := c.DeleteSavedQuery(ctx, name); err != nil {
			return dataErrMsg{Err: err}
		}

		saved, err := c.ListSavedQueries(ctx)
		if err != nil {
			return dataErrMsg{Err: err}
		}
		items := make([]queryListItem, 0, len(saved))
		for _, sq := range saved {
			items = append(items, queryListItem{name: sq.Name, cql: sq.Cql})
		}
		return queryListLoadedMsg{items: items}
	}

}

func (q *QueryList) SetItems(mode QueryListMode, items []queryListItem) {
	q.mode = mode
	q.items = items
	q.selected = 0
}
