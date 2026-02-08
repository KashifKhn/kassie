package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type ConnectedMsg struct {
	Profile string
}

type ProfileLoadedMsg struct {
	Profile string
}

type connectionErrMsg struct {
	Err error
}

type ConnectionView struct {
	theme         styles.Theme
	profiles      []string
	status        string
	selected      int
	loading       bool
	ready         bool
	spinnerFrame  int
	lastErrorTime time.Time
	retryCount    int
}

func NewConnectionView(theme styles.Theme) ConnectionView {
	return ConnectionView{
		theme:   theme,
		status:  "Fetching profiles...",
		loading: true,
	}
}

func (v ConnectionView) Init(c *client.Client) tea.Cmd {
	return tea.Batch(
		v.fetchProfilesCmd(c),
		v.loadExistingProfile(c),
		v.tickCmd(),
	)
}

func (v ConnectionView) Update(msg tea.Msg, c *client.Client, width, height int) (ConnectionView, tea.Cmd) {
	switch m := msg.(type) {
	case tickMsg:
		if v.loading {
			v.spinnerFrame = (v.spinnerFrame + 1) % 10
			return v, v.tickCmd()
		}
		return v, nil
	case tea.KeyMsg:
		switch m.String() {
		case "j", "down":
			if len(v.profiles) > 0 {
				v.selected = min(v.selected+1, len(v.profiles)-1)
			}
		case "k", "up":
			v.selected = max(v.selected-1, 0)
		case "enter":
			if v.ready && len(v.profiles) > 0 {
				profile := v.profiles[v.selected]
				v.status = fmt.Sprintf("Connecting to %s...", profile)
				v.loading = true
				v.retryCount = 0
				return v, tea.Batch(v.loginCmd(c, profile), v.tickCmd())
			}
		case "r":
			if !v.ready && !v.loading && len(v.profiles) > 0 && time.Since(v.lastErrorTime) > time.Second {
				profile := v.profiles[v.selected]
				v.status = fmt.Sprintf("Retrying connection to %s...", profile)
				v.loading = true
				v.retryCount++
				return v, tea.Batch(v.loginCmd(c, profile), v.tickCmd())
			}
		}
	case profilesMsg:
		v.loading = false
		v.ready = true
		v.profiles = m.Profiles
		if len(m.Profiles) == 0 {
			v.status = "No profiles found"
			v.ready = false
		} else {
			v.status = "Select a profile and press Enter"
		}
	case connectionErrMsg:
		v.loading = false
		v.ready = true
		v.lastErrorTime = time.Now()
		v.status = parseError(m.Err)
	case ProfileLoadedMsg:
		v.status = fmt.Sprintf("Using profile: %s", m.Profile)
		v.ready = true
	}

	return v, nil
}

func (v ConnectionView) tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (v ConnectionView) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center)

	dividerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	profileBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 2)

	selectedProfileStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("51")).
		Bold(true)

	loadingSpinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerFrame := loadingSpinnerFrames[v.spinnerFrame]

	title := titleStyle.Render("┌───────────────────────────┐\n│        K A S S I E        │\n└───────────────────────────┘")

	subtitle := subtitleStyle.Render("Database Explorer for Cassandra & ScyllaDB")

	var statusLine string
	var statusStyle lipgloss.Style

	if v.loading {
		statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Width(width - 4)
		statusLine = statusStyle.Render(fmt.Sprintf("%s %s", spinnerFrame, v.status))
	} else if !v.ready {
		statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Width(width - 4)

		errMsg := v.status
		if len(errMsg) > width-10 {
			errMsg = errMsg[:width-13] + "..."
		}
		statusLine = statusStyle.Render("✗ " + errMsg)
	} else {
		statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Width(width - 4)
		statusLine = statusStyle.Render("↓ " + v.status)
	}

	items := make([]string, 0, len(v.profiles))
	for i, p := range v.profiles {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
		if i == v.selected {
			prefix = "▶ "
			style = selectedProfileStyle
		}
		items = append(items, style.Render(prefix+p))
	}

	var profileSection string
	if len(items) == 0 {
		profileSection = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Render("No profiles configured")
	} else {
		list := lipgloss.JoinVertical(lipgloss.Left, items...)
		profileSection = profileBoxStyle.Render(list)
	}

	divider := dividerStyle.Render("─────────────────────────────")

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center)

	var helpStr string
	if !v.ready && !v.loading {
		helpStr = "j/k navigate • r retry • q quit"
		if v.retryCount > 0 {
			helpStr = fmt.Sprintf("j/k navigate • r retry (%d) • q quit", v.retryCount)
		}
	} else {
		helpStr = "j/k navigate • enter connect • q quit"
	}

	helpRendered := helpText.Render(helpStr)

	contentWidth := 50
	if width < 60 {
		contentWidth = width - 10
	}

	container := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center)

	content := container.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		title,
		"",
		subtitle,
		"",
		divider,
		"",
		profileSection,
		"",
		statusLine,
		"",
		"",
		helpRendered,
	))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

type profilesMsg struct {
	Profiles []string
}

func (v ConnectionView) fetchProfilesCmd(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		profiles, err := c.GetProfiles(ctx)
		if err != nil {
			return connectionErrMsg{Err: err}
		}

		items := make([]string, 0, len(profiles))
		for _, p := range profiles {
			items = append(items, p.Name)
		}

		return profilesMsg{Profiles: items}
	}
}

func (v ConnectionView) loginCmd(c *client.Client, profile string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := c.Login(ctx, profile)
		if err != nil {
			return connectionErrMsg{Err: err}
		}

		return ConnectedMsg{Profile: profile}
	}
}

func (v ConnectionView) loadExistingProfile(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		if c == nil {
			return nil
		}
		profile := c.Profile()
		if profile == "" {
			return nil
		}
		return ProfileLoadedMsg{Profile: profile}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func parseError(err error) string {
	if err == nil {
		return "Unknown error"
	}

	errStr := err.Error()

	if strings.Contains(errStr, "unable to discover protocol version") {
		return "Cannot connect - Check if Cassandra/ScyllaDB is running"
	}
	if strings.Contains(errStr, "connection refused") {
		return "Connection refused - Database not reachable"
	}
	if strings.Contains(errStr, "timeout") {
		return "Connection timeout - Database too slow or unreachable"
	}
	if strings.Contains(errStr, "authentication") || strings.Contains(errStr, "credentials") {
		return "Authentication failed - Check username/password"
	}
	if strings.Contains(errStr, "no hosts available") {
		return "No hosts available - Check configuration"
	}
	if strings.Contains(errStr, "profile not found") {
		return "Profile not found in config"
	}

	if len(errStr) > 80 {
		return errStr[:77] + "..."
	}

	return errStr
}
