package components

import (
	"context"
	"fmt"
	"time"

	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Sidebar struct {
	theme styles.Theme

	keyspaces []keyspaceNode
	selected  int
	scroll    int
	height    int
	loading   bool
	status    string
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
	return Sidebar{
		theme:   theme,
		loading: true,
		status:  "Loading keyspaces...",
	}
}

func (s Sidebar) Init(c *client.Client) tea.Cmd {
	return s.fetchKeyspacesCmd(c)
}

func (s Sidebar) Update(msg tea.Msg, c *client.Client) (Sidebar, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "j", "down":
			count := len(s.flatItems())
			if count > 0 {
				s.selected = minInt(s.selected+1, count-1)
			}
		case "k", "up":
			if len(s.flatItems()) > 0 {
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

	items := s.flatItems()
	if len(items) == 0 {
		items = []string{s.theme.Dim.Render("No keyspaces")}
	}

	headerLines := 2
	statusLines := 0
	if s.status != "" {
		statusLines = 1
	}
	listHeight := height - headerLines - statusLines
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

	lines := make([]string, 0, len(items)+2)
	lines = append(lines, s.theme.Header.Render("Keyspaces"))
	lines = append(lines, s.theme.Dim.Render("j/k navigate, Enter open"))
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
