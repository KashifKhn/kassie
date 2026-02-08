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
		a.state.PreviousView = a.state.View
		a.state.View = ViewHelp
		return a, nil
	case tea.KeyMsg:
		if m.String() == "?" {
			a.updateHelp()
			a.state.PreviousView = a.state.View
			a.state.View = ViewHelp
			return a, nil
		}
		if a.state.View == ViewHelp && (m.String() == "q" || m.String() == "esc" || m.String() == "?") {
			a.state.View = a.state.PreviousView
			return a, nil
		}
		if m.String() == "q" {
			return a, tea.Quit
		}
		if m.String() == "ctrl+c" {
			return a, tea.Quit
		}
	case tea.WindowSizeMsg:
		a.state = a.state.WithSize(m.Width, m.Height)
	}

	if a.state.View == ViewHelp {
		var cmd tea.Cmd
		a.help, cmd = a.help.Update(msg)
		return a, cmd
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
	titleStyle := a.theme.Title
	headerStyle := a.theme.Accent.Bold(true)
	keyStyle := a.theme.Accent
	dimStyle := a.theme.Dim

	lines := []string{
		titleStyle.Render("╔═══════════════════════════════════════════════════════════════════╗"),
		titleStyle.Render("║                      KASSIE TUI HELP                              ║"),
		titleStyle.Render("╚═══════════════════════════════════════════════════════════════════╝"),
		"",
		headerStyle.Render("━━━ NAVIGATION ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"),
		"",
		"  " + keyStyle.Render("j / k / ↓ / ↑") + "       Move cursor up/down",
		"  " + keyStyle.Render("h / l") + "               Collapse/expand keyspaces (sidebar)",
		"  " + keyStyle.Render("d / u") + "               Page down/up (10 rows)",
		"  " + keyStyle.Render("g / G") + "               Jump to top/bottom",
		"  " + keyStyle.Render("Tab") + "                 Switch to next pane",
		"  " + keyStyle.Render("Shift+Tab") + "           Switch to previous pane",
		"  " + keyStyle.Render("Ctrl+H") + "              Jump directly to sidebar",
		"  " + keyStyle.Render("Ctrl+L") + "              Jump directly to grid",
		"  " + keyStyle.Render("Ctrl+I") + "              Jump directly to inspector",
		"",
		headerStyle.Render("━━━ SEARCH & FILTER ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"),
		"",
		"  " + keyStyle.Render("Ctrl+F") + "              Search in sidebar (fuzzy) or grid (text)",
		"  " + keyStyle.Render("/") + "                   Open WHERE filter (grid) or search (sidebar)",
		"  " + keyStyle.Render("s") + "                   Toggle system keyspaces visibility",
		"  " + keyStyle.Render("n / N") + "               Next/Previous search match",
		"  " + keyStyle.Render("Enter") + "               Confirm search/filter",
		"  " + keyStyle.Render("Esc") + "                 Cancel search/filter",
		"",
		headerStyle.Render("━━━ ACTIONS ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"),
		"",
		"  " + keyStyle.Render("Enter") + "               Select item / Expand keyspace",
		"  " + keyStyle.Render("r") + "                   Refresh current view",
		"  " + keyStyle.Render("n") + "                   Load next page (if available)",
		"  " + keyStyle.Render("Ctrl+C") + "              Copy JSON to clipboard (inspector)",
		"  " + keyStyle.Render("?") + "                   Show/Hide this help screen",
		"  " + keyStyle.Render("q") + "                   Back / Quit application",
		"",
		headerStyle.Render("━━━ FEATURES ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"),
		"",
		"  • " + dimStyle.Render("Fuzzy Search") + "       - fzf-powered search with scoring",
		"  • " + dimStyle.Render("Virtual Scrolling") + "  - Smooth performance with 1000+ rows",
		"  • " + dimStyle.Render("Schema Caching") + "    - 10-minute TTL cache for schemas",
		"  • " + dimStyle.Render("Client-side Search") + " - Fast text search in current results",
		"  • " + dimStyle.Render("System Toggle") + "     - Hide/show Cassandra system keyspaces",
		"  • " + dimStyle.Render("Filter Validation") + " - Prevents dangerous SQL operations",
		"",
		headerStyle.Render("━━━ TIPS ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"),
		"",
		"  • Fuzzy search: Type 'bum' to match 'business_members'",
		"  • Press 's' to hide system_* keyspaces for cleaner view",
		"  • Use Ctrl+F in grid to search across all columns",
		"  • Schemas are cached - switching tables is instant!",
		"  • WHERE filters support: =, >, <, >=, <=, IN, CONTAINS, AND, OR",
		"",
		dimStyle.Render("                    Press ? or Esc or q to close"),
	}

	a.help = views.NewHelpView(strings.Join(lines, "\n"))
}
