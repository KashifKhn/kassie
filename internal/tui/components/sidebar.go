package components

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Sidebar struct {
	theme styles.Theme

	keyspaces    []keyspaceNode
	selected     int
	scroll       int
	height       int
	loading      bool
	status       string
	searchInput  textinput.Model
	searchActive bool
	searchQuery  string
}

type keyspaceNode struct {
	name     string
	expanded bool
	tables   []string
}

type KeyspaceSelectedMsg struct {
	Keyspace string
}

type TableSelectedMsg struct {
	Keyspace string
	Table    string
}

type sidebarErrMsg struct {
	Err error
}

type keyspacesMsg struct {
	Keyspaces []string
}

type tablesMsg struct {
	Keyspace string
	Tables   []string
}

func NewSidebar(theme styles.Theme) Sidebar {
	input := textinput.New()
	input.Placeholder = "Search tables..."
	input.Prompt = "🔍 "
	input.CharLimit = 100
	input.Width = 30

	return Sidebar{
		theme:       theme,
		loading:     true,
		status:      "Loading keyspaces...",
		searchInput: input,
	}
}

func (s Sidebar) Init(c *client.Client) tea.Cmd {
	return s.fetchKeyspacesCmd(c)
}

func (s Sidebar) Update(msg tea.Msg, c *client.Client) (Sidebar, tea.Cmd) {
	var cmd tea.Cmd

	switch m := msg.(type) {
	case tea.KeyMsg:
		if s.searchActive {
			switch m.String() {
			case "esc":
				s.searchActive = false
				s.searchQuery = ""
				s.searchInput.SetValue("")
				s.searchInput.Blur()
				return s, nil
			case "enter":
				s.searchActive = false
				s.searchQuery = strings.TrimSpace(s.searchInput.Value())
				s.searchInput.Blur()
				s.selected = 0
				return s, nil
			default:
				s.searchInput, cmd = s.searchInput.Update(msg)
				return s, cmd
			}
		}

		switch m.String() {
		case "ctrl+f", "/":
			s.searchActive = true
			s.searchInput.Focus()
			return s, nil
		case "esc":
			if s.searchQuery != "" {
				s.searchQuery = ""
				s.searchInput.SetValue("")
				s.selected = 0
				return s, nil
			}
		case "j", "down":
			count := len(s.filteredItems())
			if count > 0 {
				s.selected = minInt(s.selected+1, count-1)
			}
		case "k", "up":
			if len(s.filteredItems()) > 0 {
				s.selected = maxInt(s.selected-1, 0)
			}
		case "enter":
			return s.handleSelect(c)
		case "l", "right":
			return s.expandSelected(c)
		case "h", "left":
			return s.collapseSelected()
		}
	case keyspacesMsg:
		s.loading = false
		s.status = ""
		s.keyspaces = make([]keyspaceNode, 0, len(m.Keyspaces))
		for _, ks := range m.Keyspaces {
			s.keyspaces = append(s.keyspaces, keyspaceNode{name: ks})
		}
	case tablesMsg:
		s.applyTables(m.Keyspace, m.Tables)
	case sidebarErrMsg:
		s.loading = false
		s.status = fmt.Sprintf("Error: %s", m.Err)
	}

	return s, nil
}

func (s *Sidebar) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	s.height = height

	items := s.filteredItems()
	if len(items) == 0 {
		if s.searchQuery != "" {
			items = []string{s.theme.Dim.Render("No matches")}
		} else {
			items = []string{s.theme.Dim.Render("No keyspaces")}
		}
	}

	headerLines := 2
	statusLines := 0
	if s.status != "" {
		statusLines = 1
	}
	searchLines := 0
	if s.searchActive {
		searchLines = 1
	}
	if s.searchQuery != "" && !s.searchActive {
		searchLines = 1
	}

	listHeight := height - headerLines - statusLines - searchLines
	if listHeight < 0 {
		listHeight = 0
	}

	if len(items) > listHeight {
		s.scroll = clampInt(s.scroll, 0, len(items)-listHeight)
		if s.selected < s.scroll {
			s.scroll = s.selected
		} else if s.selected >= s.scroll+listHeight {
			s.scroll = s.selected - listHeight + 1
		}
		items = items[s.scroll:minInt(s.scroll+listHeight, len(items))]
	}

	lines := make([]string, 0, len(items)+4)
	lines = append(lines, s.theme.Header.Render("Keyspaces"))

	helpText := "j/k navigate, Enter open"
	if !s.searchActive {
		helpText += ", / search"
	}
	lines = append(lines, s.theme.Dim.Render(helpText))

	if s.searchActive {
		s.searchInput.Width = width - 4
		searchBar := s.searchInput.View()
		searchStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("51")).
			Padding(0, 1)
		lines = append(lines, searchStyle.Render(searchBar))
	} else if s.searchQuery != "" {
		queryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Italic(true)
		clearHint := s.theme.Dim.Render(" (Esc to clear)")
		lines = append(lines, queryStyle.Render("🔍 "+s.searchQuery)+clearHint)
	}

	lines = append(lines, items...)

	if s.status != "" {
		lines = append(lines, "", s.theme.Status.Render(s.status))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.NewStyle().Width(width).Height(height).Render(content)
}

func (s Sidebar) handleSelect(c *client.Client) (Sidebar, tea.Cmd) {
	ksIndex, tblIndex, ok := s.selectedIndex()
	if !ok {
		return s, nil
	}

	if tblIndex >= 0 {
		keyspace := s.keyspaces[ksIndex].name
		table := s.keyspaces[ksIndex].tables[tblIndex]
		return s, func() tea.Msg { return TableSelectedMsg{Keyspace: keyspace, Table: table} }
	}

	keyspace := s.keyspaces[ksIndex].name
	return s.toggleKeyspace(c, ksIndex, keyspace)
}

func (s Sidebar) expandSelected(c *client.Client) (Sidebar, tea.Cmd) {
	ksIndex, tblIndex, ok := s.selectedIndex()
	if !ok || tblIndex >= 0 {
		return s, nil
	}
	if s.keyspaces[ksIndex].expanded {
		return s, nil
	}
	keyspace := s.keyspaces[ksIndex].name
	return s.toggleKeyspace(c, ksIndex, keyspace)
}

func (s Sidebar) collapseSelected() (Sidebar, tea.Cmd) {
	ksIndex, tblIndex, ok := s.selectedIndex()
	if !ok {
		return s, nil
	}
	if tblIndex >= 0 {
		s.selected = s.keyspaceOffset(ksIndex)
		return s, nil
	}
	if s.keyspaces[ksIndex].expanded {
		s.keyspaces[ksIndex].expanded = false
	}
	return s, nil
}

func (s Sidebar) toggleKeyspace(c *client.Client, ksIndex int, keyspace string) (Sidebar, tea.Cmd) {
	ks := s.keyspaces[ksIndex]
	if ks.expanded {
		s.keyspaces[ksIndex].expanded = false
		return s, nil
	}

	s.keyspaces[ksIndex].expanded = true
	if len(ks.tables) > 0 {
		return s, func() tea.Msg { return KeyspaceSelectedMsg{Keyspace: keyspace} }
	}

	return s, tea.Batch(
		s.fetchTablesCmd(c, ks.name),
		func() tea.Msg { return KeyspaceSelectedMsg{Keyspace: keyspace} },
	)
}

func (s Sidebar) fetchKeyspacesCmd(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		keyspaces, err := c.ListKeyspaces(ctx)
		if err != nil {
			return sidebarErrMsg{Err: err}
		}

		items := make([]string, 0, len(keyspaces))
		for _, ks := range keyspaces {
			items = append(items, ks.Name)
		}

		return keyspacesMsg{Keyspaces: items}
	}
}

func (s Sidebar) fetchTablesCmd(c *client.Client, keyspace string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tables, err := c.ListTables(ctx, keyspace)
		if err != nil {
			return sidebarErrMsg{Err: err}
		}

		items := make([]string, 0, len(tables))
		for _, tbl := range tables {
			items = append(items, tbl.Name)
		}

		return tablesMsg{Keyspace: keyspace, Tables: items}
	}
}

func (s Sidebar) applyTables(keyspace string, tables []string) {
	for i := range s.keyspaces {
		if s.keyspaces[i].name == keyspace {
			s.keyspaces[i].tables = tables
			return
		}
	}
}

func (s Sidebar) flatItems() []string {
	items := make([]string, 0)
	selectedStyle := s.theme.Selected
	for _, ks := range s.keyspaces {
		prefix := "> "
		if ks.expanded {
			prefix = "v "
		}
		items = append(items, s.theme.SidebarKey.Render(prefix+ks.name))
		if ks.expanded {
			for _, tbl := range ks.tables {
				items = append(items, s.theme.SidebarTbl.Render("    * "+tbl))
			}
		}
	}

	for i := range items {
		if i == s.selected {
			items[i] = selectedStyle.Render(items[i])
		}
	}

	return items
}

func (s Sidebar) filteredItems() []string {
	if s.searchQuery == "" {
		return s.flatItems()
	}

	query := strings.ToLower(s.searchQuery)
	items := make([]string, 0)
	selectedStyle := s.theme.Selected

	itemIndex := 0
	for _, ks := range s.keyspaces {
		keyspaceMatches := strings.Contains(strings.ToLower(ks.name), query)
		tableMatches := false

		if ks.expanded || keyspaceMatches {
			for _, tbl := range ks.tables {
				if strings.Contains(strings.ToLower(tbl), query) {
					tableMatches = true
					break
				}
			}
		}

		if keyspaceMatches || tableMatches {
			prefix := "> "
			if ks.expanded {
				prefix = "v "
			}

			ksText := prefix + ks.name
			if keyspaceMatches {
				ksText = s.highlightMatch(ksText, query)
			}
			ksRendered := s.theme.SidebarKey.Render(ksText)

			if itemIndex == s.selected {
				ksRendered = selectedStyle.Render(ksRendered)
			}
			items = append(items, ksRendered)
			itemIndex++

			if ks.expanded {
				for _, tbl := range ks.tables {
					tblMatches := strings.Contains(strings.ToLower(tbl), query)
					if s.searchQuery == "" || tblMatches || keyspaceMatches {
						tblText := "    * " + tbl
						if tblMatches {
							tblText = "    * " + s.highlightMatch(tbl, query)
						}
						tblRendered := s.theme.SidebarTbl.Render(tblText)

						if itemIndex == s.selected {
							tblRendered = selectedStyle.Render(tblRendered)
						}
						items = append(items, tblRendered)
						itemIndex++
					}
				}
			}
		}
	}

	return items
}

func (s Sidebar) highlightMatch(text, query string) string {
	lower := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)
	idx := strings.Index(lower, lowerQuery)

	if idx == -1 {
		return text
	}

	matchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	before := text[:idx]
	match := text[idx : idx+len(query)]
	after := text[idx+len(query):]

	return before + matchStyle.Render(match) + after
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (s Sidebar) selectedIndex() (int, int, bool) {
	index := s.selected
	for ksIndex, ks := range s.keyspaces {
		if index == 0 {
			return ksIndex, -1, true
		}
		index--
		if ks.expanded {
			if index < len(ks.tables) {
				return ksIndex, index, true
			}
			index -= len(ks.tables)
		}
	}
	return 0, 0, false
}

func (s Sidebar) keyspaceOffset(ksIndex int) int {
	offset := 0
	for i, ks := range s.keyspaces {
		if i == ksIndex {
			return offset
		}
		offset++
		if ks.expanded {
			offset += len(ks.tables)
		}
	}
	return 0
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
