package tui

import (
	"fmt"
	"strings"

	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	"github.com/KashifKhn/kassie/internal/tui/views"
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	client *client.Client
	state  AppState
	theme  styles.Theme

	connection views.ConnectionView
	explorer   views.ExplorerView
	help       views.HelpView
}

func NewApp(client *client.Client) *App {
	theme := styles.DefaultTheme()
	return &App{
		client:     client,
		state:      NewState(),
		theme:      theme,
		connection: views.NewConnectionView(theme),
		explorer:   views.NewExplorerView(theme),
		help:       views.NewHelpView(""),
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.connection.Init(a.client),
		a.explorer.Init(a.client),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case views.ConnectedMsg:
		a.state.Profile = m.Profile
		a.state.Status = "Connected"
		a.state.View = ViewExplorer
		a.explorer.SetProfile(m.Profile)
		var cmd tea.Cmd
		a.explorer, cmd = a.explorer.Reload(a.client)
		return a, cmd
	case views.ProfileLoadedMsg:
		a.state.Profile = m.Profile
		a.state.View = ViewExplorer
		a.explorer.SetProfile(m.Profile)
		var cmd tea.Cmd
		a.explorer, cmd = a.explorer.Reload(a.client)
		return a, cmd
	case views.ShowHelpMsg:
		a.updateHelp()
		a.state.View = ViewHelp
		return a, nil
	case tea.KeyMsg:
		if a.state.View == ViewHelp && (m.String() == "q" || m.String() == "esc" || m.String() == "?") {
			a.state.View = ViewExplorer
			return a, nil
		}
		if m.String() == "q" && a.state.View == ViewExplorer {
			return a, tea.Quit
		}
	case tea.WindowSizeMsg:
		a.state = a.state.WithSize(m.Width, m.Height)
	}

	switch m := msg.(type) {
	case tea.KeyMsg:
		if m.String() == "ctrl+c" || m.String() == "q" {
			return a, tea.Quit
		}
	}

	if a.state.View == ViewConnection {
		var cmd tea.Cmd
		a.connection, cmd = a.connection.Update(msg, a.client, a.state.Width, a.state.Height)
		return a, cmd
	}

	if a.state.View == ViewExplorer {
		var cmd tea.Cmd
		a.explorer, cmd = a.explorer.Update(msg, a.client)
		return a, cmd
	}

	return a, nil
}

func (a *App) View() string {
	switch a.state.View {
	case ViewConnection:
		return a.connection.View(a.state.Width, a.state.Height)
	case ViewExplorer:
		return a.explorer.View(a.state.Width, a.state.Height)
	case ViewHelp:
		return a.help.View(a.state.Width, a.state.Height)
	default:
		return a.renderPlaceholder("Unknown view")
	}
}

func (a *App) renderPlaceholder(text string) string {
	return a.theme.Dim.Render(fmt.Sprintf("%s (coming soon)", text))
}

func (a *App) updateHelp() {
	lines := []string{
		a.theme.Title.Render("Kassie TUI Help"),
		"",
		"Navigation:",
		"  j/k or Up/Down  Move up/down",
		"  h/l             Collapse/expand",
		"  Tab             Next pane",
		"  Shift+Tab       Previous pane",
		"",
		"Actions:",
		"  Enter           Select",
		"  /               Filter",
		"  n               Next page",
		"  r               Refresh",
		"  Ctrl+H/L/I      Jump to sidebar/grid/inspector",
		"  ?               Help",
		"  q               Back/Quit",
	}

	a.help = views.NewHelpView(strings.Join(lines, "\n"))
}
