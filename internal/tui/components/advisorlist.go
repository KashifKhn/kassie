package components

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AdvisorClosedMsg struct{}

type AdvisorList struct {
	theme    styles.Theme
	active   bool
	keyspace string
	findings []*pb.AdvisorFinding
	tables   int32
	selected int
	err      string
}

func NewAdvisorList(theme styles.Theme) AdvisorList {
	return AdvisorList{theme: theme}
}

func (a AdvisorList) IsActive() bool {
	return a.active
}

func (a AdvisorList) Activate(cl *client.Client, keyspace string) (AdvisorList, tea.Cmd) {
	a.active = true
	a.keyspace = keyspace
	a.selected = 0
	a.err = ""
	a.findings = nil
	if keyspace == "" {
		a.err = "no keyspace selected"
		return a, nil
	}
	return a, fetchAdvisorCmd(cl, keyspace)
}

func (a AdvisorList) Deactivate() AdvisorList {
	a.active = false
	a.findings = nil
	a.err = ""
	return a
}

func (a AdvisorList) Update(msg tea.Msg, cl *client.Client) (AdvisorList, tea.Cmd) {
	if !a.active {
		return a, nil
	}

	switch m := msg.(type) {
	case advisorLoadedMsg:
		a.findings = m.Findings
		a.tables = m.Tables
		a.err = ""
		return a, nil
	case dataErrMsg:
		a.err = m.Err.Error()
		return a, nil
	}

	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	switch key.String() {
	case "j", "down":
		if a.selected < len(a.findings)-1 {
			a.selected++
		}
	case "k", "up":
		if a.selected > 0 {
			a.selected--
		}
	case "r":
		return a, fetchAdvisorCmd(cl, a.keyspace)
	case "esc", "ctrl+a":
		a = a.Deactivate()
		return a, func() tea.Msg { return AdvisorClosedMsg{} }
	}

	return a, nil
}

func (a AdvisorList) View(width int) string {
	if !a.active {
		return ""
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	selStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	localGreen := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	remStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))

	lines := []string{titleStyle.Render(fmt.Sprintf("Advisor — %s (%d tables)", a.keyspace, a.tables))}

	if a.err != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗ "+a.err))
	}

	if len(a.findings) == 0 && a.err == "" {
		lines = append(lines, localGreen.Render("✓ no findings"))
	}

	maxWidth := width - 6
	for i, f := range a.findings {
		style := dimStyle
		if f.Severity == "warning" {
			style = warnStyle
		} else {
			style = infoStyle
		}

		head := fmt.Sprintf("%s [%s] %s", strings.ToUpper(f.Severity), f.Rule, f.Table)
		if len(head) > maxWidth {
			head = head[:maxWidth-3] + "..."
		}

		if i == a.selected {
			lines = append(lines, selStyle.Render("▸ "+style.Render(head)))
		} else {
			lines = append(lines, "  "+style.Render(head))
		}

		msg := f.Message
		for len(msg) > 0 {
			chunk := msg
			if len(chunk) > maxWidth {
				chunk = chunk[:maxWidth]
			}
			lines = append(lines, "    "+dimStyle.Render(chunk))
			msg = msg[len(chunk):]
		}

		if f.Remediation != "" {
			rem := f.Remediation
			for len(rem) > 0 {
				chunk := rem
				if len(chunk) > maxWidth {
					chunk = chunk[:maxWidth]
				}
				lines = append(lines, "    "+remStyle.Render(chunk))
				rem = rem[len(chunk):]
			}
		}
		lines = append(lines, "")
	}

	lines = append(lines, dimStyle.Render("j/k navigate • r refresh • esc close"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")).
		Padding(0, 1).
		Width(width - 2).
		Render(strings.Join(lines, "\n"))
}

func fetchAdvisorCmd(c *client.Client, keyspace string) tea.Cmd {
	return func() tea.Msg {
		if c == nil {
			return dataErrMsg{Err: errNoClient}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		findings, tables, err := c.AnalyzeKeyspace(ctx, keyspace)
		if err != nil {
			return dataErrMsg{Err: err}
		}
		return advisorLoadedMsg{Findings: findings, Tables: tables}
	}
}

type advisorLoadedMsg struct {
	Findings []*pb.AdvisorFinding
	Tables   int32
}
