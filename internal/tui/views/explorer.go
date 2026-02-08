package views

import (
	"fmt"
	"time"

	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/tui/cache"
	"github.com/KashifKhn/kassie/internal/tui/components"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CopySuccessMsg struct{}

type CopyErrorMsg struct {
	Err error
}

type ExplorerView struct {
	theme styles.Theme

	sidebar     components.Sidebar
	grid        components.DataGrid
	inspect     components.Inspector
	filter      components.FilterBar
	status      components.StatusBar
	active      pane
	profile     string
	message     string
	schemaCache *cache.SchemaCache
}

type pane int

const (
	paneSidebar pane = iota
	paneGrid
	paneInspector
)

func NewExplorerView(theme styles.Theme) ExplorerView {
	schemaCache := cache.NewSchemaCache(10 * time.Minute)
	return ExplorerView{
		theme:       theme,
		sidebar:     components.NewSidebar(theme),
		grid:        components.NewDataGrid(theme, schemaCache),
		inspect:     components.NewInspector(theme),
		filter:      components.NewFilterBar(theme),
		status:      components.NewStatusBar(theme),
		active:      paneSidebar,
		schemaCache: schemaCache,
	}
}

func (v ExplorerView) Reload(c *client.Client) (ExplorerView, tea.Cmd) {
	v.sidebar = components.NewSidebar(v.theme)
	v.grid = components.NewDataGrid(v.theme, v.schemaCache)
	v.inspect = components.NewInspector(v.theme)
	v.filter = components.NewFilterBar(v.theme)
	v.active = paneSidebar

	return v, tea.Batch(
		v.sidebar.Init(c),
		v.grid.Init(),
	)
}

func (v ExplorerView) Init(c *client.Client) tea.Cmd {
	return tea.Batch(
		v.sidebar.Init(c),
		v.grid.Init(),
	)
}

func (v ExplorerView) Update(msg tea.Msg, c *client.Client) (ExplorerView, tea.Cmd) {
	switch m := msg.(type) {
	case CopySuccessMsg:
		v.message = "✓ Copied to clipboard"
		return v, v.clearMessageAfter(2)
	case CopyErrorMsg:
		v.message = fmt.Sprintf("✗ Copy failed: %v", m.Err)
		return v, v.clearMessageAfter(3)
	case clearMessageMsg:
		v.message = ""
		return v, nil
	case components.TableSelectedMsg:
		var cmd tea.Cmd
		v.grid, cmd = v.grid.LoadTable(c, m.Keyspace, m.Table)
		return v, cmd
	case components.KeyspaceSelectedMsg:
		v.filter = v.filter.Deactivate()
		return v, nil
	case components.RowSelectedMsg:
		v.inspect.SetRow(m.Row)
		return v, nil
	case components.FilterAppliedMsg:
		v.filter = v.filter.Deactivate()
		var cmd tea.Cmd
		v.grid, cmd = v.grid.ApplyFilter(c, m.Where)
		return v, cmd
	case components.FilterCanceledMsg:
		v.filter = v.filter.Deactivate()
		return v, nil
	case tea.KeyMsg:
		if m.String() == "r" {
			var cmd tea.Cmd
			v.grid, cmd = v.grid.Refresh(c)
			return v, cmd
		}
		if m.String() == "q" {
			return v, nil
		}
	}

	if v.filter.IsActive() {
		var cmd tea.Cmd
		v.filter, cmd = v.filter.Update(msg)
		return v, cmd
	}

	if _, ok := msg.(tea.KeyMsg); !ok {
		var cmd tea.Cmd
		var cmd2 tea.Cmd
		v.sidebar, cmd = v.sidebar.Update(msg, c)
		v.grid, cmd2 = v.grid.Update(msg, c)
		return v, tea.Batch(cmd, cmd2)
	}

	var cmd tea.Cmd

	switch v.active {
	case paneSidebar:
		v.sidebar, cmd = v.sidebar.Update(msg, c)
	case paneGrid:
		v.grid, cmd = v.grid.Update(msg, c)
	case paneInspector:
		keyMsg, ok := msg.(tea.KeyMsg)
		if ok {
			switch keyMsg.String() {
			case "j", "down":
				v.inspect.ScrollDown()
			case "k", "up":
				v.inspect.ScrollUp()
			case "d":
				height := 20
				v.inspect.PageDown(height)
			case "u":
				height := 20
				v.inspect.PageUp(height)
			case "ctrl+c":
				return v, v.copyToClipboardCmd()
			}
		}
	}

	v, cmd = v.handleNavigation(msg, cmd)
	return v, cmd
}

func (v ExplorerView) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	if width < 60 || height < 10 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, v.theme.Dim.Render("Resize terminal"))
	}

	leftWidth := maxInt(24, width/4)
	middleWidth := maxInt(30, width/2)
	rightWidth := width - leftWidth - middleWidth
	if rightWidth < 20 {
		middleWidth = maxInt(30, width-leftWidth-20)
		rightWidth = width - leftWidth - middleWidth
	}

	border := v.theme.Panel
	filterHeight := 0
	filterView := ""
	if v.filter.IsActive() {
		filterHeight = 3
		filterView = v.filter.View(width)
	}

	contentHeight := height - filterHeight - 1
	if contentHeight < 1 {
		contentHeight = height
	}

	leftBorder := border
	middleBorder := border
	rightBorder := border
	if v.active == paneSidebar {
		leftBorder = v.theme.PanelHot
	}
	if v.active == paneGrid {
		middleBorder = v.theme.PanelHot
	}
	if v.active == paneInspector {
		rightBorder = v.theme.PanelHot
	}

	left := leftBorder.Width(leftWidth).Height(contentHeight).Render(v.sidebar.View(leftWidth-2, contentHeight-2))
	middle := middleBorder.Width(middleWidth).Height(contentHeight).Render(v.grid.View(middleWidth-2, contentHeight-2))
	right := rightBorder.Width(rightWidth).Height(contentHeight).Render(v.inspect.View(rightWidth-2, contentHeight-2))

	row := lipgloss.JoinHorizontal(lipgloss.Top, left, middle, right)
	statusText := v.grid.Status()
	if statusText == "" {
		statusText = "Ready"
	}
	if v.message != "" {
		statusText = v.message
	}
	statusHint := fmt.Sprintf("Pane: %s | Tab switch", v.paneLabel())
	status := v.status.View(width, v.profile, v.grid.Keyspace(), v.grid.Table(), statusText+" | "+statusHint)

	parts := []string{row}
	if filterView != "" {
		parts = append(parts, filterView)
	}
	parts = append(parts, status)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (v ExplorerView) handleNavigation(msg tea.Msg, cmd tea.Cmd) (ExplorerView, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return v, cmd
	}

	switch key.String() {
	case "tab":
		v.active = (v.active + 1) % 3
	case "shift+tab":
		v.active = (v.active + 2) % 3
	case "ctrl+h":
		v.active = paneSidebar
	case "ctrl+l":
		v.active = paneGrid
	case "ctrl+i":
		v.active = paneInspector
	case "/":
		if v.active == paneSidebar {
			v.sidebar, cmd = v.sidebar.ActivateSearch()
		} else if v.active == paneGrid {
			v.filter = v.filter.Activate(v.grid.Filter())
		}
	case "?":
		return v, tea.Batch(cmd, func() tea.Msg { return ShowHelpMsg{} })
	}

	return v, cmd
}

func (v ExplorerView) paneLabel() string {
	switch v.active {
	case paneSidebar:
		return "sidebar"
	case paneGrid:
		return "grid"
	case paneInspector:
		return "inspector"
	default:
		return ""
	}
}

func (v *ExplorerView) SetProfile(profile string) {
	v.profile = profile
}

type ShowHelpMsg struct{}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type clearMessageMsg struct{}

func (v ExplorerView) clearMessageAfter(seconds int) tea.Cmd {
	return tea.Tick(time.Duration(seconds)*time.Second, func(t time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

func (v ExplorerView) copyToClipboardCmd() tea.Cmd {
	return func() tea.Msg {
		if err := v.inspect.CopyToClipboard(); err != nil {
			return CopyErrorMsg{Err: err}
		}
		return CopySuccessMsg{}
	}
}
