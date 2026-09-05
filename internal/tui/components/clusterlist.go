package components

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ClusterClosedMsg struct{}

type ClusterList struct {
	theme    styles.Theme
	active   bool
	nodes    []*pb.ClusterNode
	selected int
	err      string
}

func NewClusterList(theme styles.Theme) ClusterList {
	return ClusterList{theme: theme}
}

func (c ClusterList) IsActive() bool {
	return c.active
}

func (c ClusterList) Activate(cl *client.Client) (ClusterList, tea.Cmd) {
	c.active = true
	c.selected = 0
	c.err = ""
	c.nodes = nil
	return c, fetchClusterInfoCmd(cl)
}

func (c ClusterList) Deactivate() ClusterList {
	c.active = false
	c.nodes = nil
	c.err = ""
	return c
}

func (c ClusterList) Update(msg tea.Msg, cl *client.Client) (ClusterList, tea.Cmd) {
	if !c.active {
		return c, nil
	}

	switch m := msg.(type) {
	case clusterLoadedMsg:
		c.nodes = m.Nodes
		c.err = ""
		return c, nil
	case dataErrMsg:
		c.err = m.Err.Error()
		return c, nil
	}

	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return c, nil
	}

	switch key.String() {
	case "j", "down":
		if c.selected < len(c.nodes)-1 {
			c.selected++
		}
	case "k", "up":
		if c.selected > 0 {
			c.selected--
		}
	case "r":
		return c, fetchClusterInfoCmd(cl)
	case "esc", "ctrl+r":
		c = c.Deactivate()
		return c, func() tea.Msg { return ClusterClosedMsg{} }
	}

	return c, nil
}

func (c ClusterList) View(width int) string {
	if !c.active {
		return ""
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	selStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	localStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))

	lines := []string{accentStyle.Bold(true).Render(fmt.Sprintf("Cluster — %d nodes", len(c.nodes)))}

	if c.err != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗ "+c.err))
	}

	grouped := map[string][]*pb.ClusterNode{}
	for _, node := range c.nodes {
		key := node.DataCenter
		if key == "" {
			key = "unknown"
		}
		grouped[key] = append(grouped[key], node)
	}

	dcs := make([]string, 0, len(grouped))
	for dc := range grouped {
		dcs = append(dcs, dc)
	}
	sort.Strings(dcs)

	idx := 0
	for _, dc := range dcs {
		lines = append(lines, "", dimStyle.Render(fmt.Sprintf("datacenter: %s (%d)", dc, len(grouped[dc]))))
		for _, node := range grouped[dc] {
			status := node.Status
			if status == "" {
				status = "up"
			}
			entry := fmt.Sprintf("%-15s %-6s %-6s v%-8s %3d tokens  %s",
				node.Address, node.Rack, status, node.ReleaseVersion, node.TokenCount, marker(node.Local))

			if idx == c.selected {
				lines = append(lines, selStyle.Render("▸ "+entry))
			} else if node.Local {
				lines = append(lines, localStyle.Render("  "+entry))
			} else {
				lines = append(lines, dimStyle.Render("  "+entry))
			}
			idx++
		}
	}

	lines = append(lines, "", dimStyle.Render("j/k navigate • r refresh • esc close"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")).
		Padding(0, 1).
		Width(width - 2).
		Render(strings.Join(lines, "\n"))
}

func marker(local bool) string {
	if local {
		return "(coordinator)"
	}
	return ""
}

type clusterLoadedMsg struct {
	Nodes []*pb.ClusterNode
}

func fetchClusterInfoCmd(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		if c == nil {
			return dataErrMsg{Err: errNoClient}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		nodes, err := c.GetClusterInfo(ctx)
		if err != nil {
			return dataErrMsg{Err: err}
		}
		return clusterLoadedMsg{Nodes: nodes}
	}
}
