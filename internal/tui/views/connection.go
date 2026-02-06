package views

import (
	"context"
	"fmt"
	"time"

	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
	theme    styles.Theme
	profiles []string
	status   string
	selected int
	loading  bool
	ready    bool
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
	)
}

func (v ConnectionView) Update(msg tea.Msg, c *client.Client, width, height int) (ConnectionView, tea.Cmd) {
	switch m := msg.(type) {
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
				return v, v.loginCmd(c, profile)
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
		v.ready = false
		v.status = fmt.Sprintf("Error: %s", m.Err)
	case ProfileLoadedMsg:
		v.status = fmt.Sprintf("Using profile: %s", m.Profile)
		v.ready = true
	}

	return v, nil
}

func (v ConnectionView) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	items := make([]string, 0, len(v.profiles))
	for i, p := range v.profiles {
		label := p
		if i == v.selected {
			label = v.theme.Accent.Render("> " + label)
		} else {
			label = "  " + label
		}
		items = append(items, label)
	}

	if len(items) == 0 {
		items = append(items, v.theme.Dim.Render("No profiles"))
	}

	header := v.theme.Title.Render("Kassie")
	list := lipgloss.JoinVertical(lipgloss.Left, items...)
	status := v.theme.Status.Render(v.status)

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", list, "", status)
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
