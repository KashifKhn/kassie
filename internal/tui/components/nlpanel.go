package components

import (
	"context"
	"time"

	"github.com/KashifKhn/kassie/internal/client"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NlTranslatedMsg struct {
	CQL      string
	Provider string
	Question string
	Keyspace string
}

type NlCanceledMsg struct{}

type nlErrMsg struct {
	Err error
}

type NlProvidersLoadedMsg struct {
	Available []string
	Default   string
}

var nlProviderCycle = []string{"auto", "opencode", "claude", "codex", "gemini"}

type NlPanel struct {
	active       bool
	input        []rune
	loading      bool
	err          string
	translated   *NlTranslatedMsg
	keyspace     string
	table        string
	provider     string
	providerIdx  int
	providers    []string
	model        string
	editingModel bool
	modelInput   []rune
}

func NewNlPanel() NlPanel {
	return NlPanel{}
}

func (n NlPanel) IsActive() bool {
	return n.active
}

func (n NlPanel) Activate(keyspace, table string) (NlPanel, tea.Cmd) {
	n.active = true
	n.input = nil
	n.loading = false
	n.err = ""
	n.translated = nil
	n.keyspace = keyspace
	n.table = table
	n.provider = "auto"
	n.providerIdx = 0
	return n, fetchNlProvidersCmd
}

func (n NlPanel) Deactivate() NlPanel {
	n.active = false
	n.input = nil
	n.loading = false
	n.err = ""
	n.translated = nil
	return n
}

func (n NlPanel) Provider() string {
	return n.provider
}

func (n NlPanel) Model() string {
	return n.model
}

func fetchNlProvidersCmd() tea.Msg {
	return NlProvidersLoadedMsg{
		Available: nlProviderCycle,
		Default:   "auto",
	}
}

func (n NlPanel) Update(msg tea.Msg, c *client.Client) (NlPanel, tea.Cmd) {
	if !n.active {
		return n, nil
	}

	switch m := msg.(type) {
	case nlErrMsg:
		n.loading = false
		n.err = m.Err.Error()
		return n, nil
	case NlTranslatedMsg:
		n.loading = false
		n.translated = &m
		return n, nil
	case NlProvidersLoadedMsg:
		n.providers = m.Available
		return n, nil
	case tea.KeyMsg:
		if n.loading {
			if m.String() == "esc" || m.String() == "ctrl+c" {
				n = n.Deactivate()
				return n, func() tea.Msg { return NlCanceledMsg{} }
			}
			return n, nil
		}

		if n.translated != nil {
			switch m.String() {
			case "enter":
				msg := *n.translated
				n = n.Deactivate()
				return n, func() tea.Msg { return msg }
			case "esc", "ctrl+n":
				n = n.Deactivate()
				return n, func() tea.Msg { return NlCanceledMsg{} }
			}
			return n, nil
		}

		if n.editingModel {
			switch m.String() {
			case "enter":
				n.model = string(n.modelInput)
				n.editingModel = false
				return n, nil
			case "esc":
				n.editingModel = false
				n.modelInput = []rune(n.model)
				return n, nil
			case "backspace":
				if len(n.modelInput) > 0 {
					n.modelInput = n.modelInput[:len(n.modelInput)-1]
				}
				return n, nil
			default:
				if len(m.Runes) > 0 {
					n.modelInput = append(n.modelInput, m.Runes...)
				}
				return n, nil
			}
		}

		switch m.String() {
		case "ctrl+p":
			n.providerIdx = (n.providerIdx + 1) % len(nlProviderCycle)
			n.provider = nlProviderCycle[n.providerIdx]
			return n, nil
		case "ctrl+e":
			n.editingModel = true
			n.modelInput = []rune(n.model)
			return n, nil
		case "enter":
			if len(n.input) == 0 {
				return n, nil
			}
			question := string(n.input)
			keyspace := n.keyspace
			table := n.table
			provider := n.provider
			model := n.model
			n.loading = true
			n.err = ""
			return n, translateNlCmd(c, question, keyspace, table, provider, model)
		case "backspace":
			if len(n.input) > 0 {
				n.input = n.input[:len(n.input)-1]
			}
		case "esc", "ctrl+n":
			n = n.Deactivate()
			return n, func() tea.Msg { return NlCanceledMsg{} }
		default:
			if len(m.Runes) > 0 {
				n.input = append(n.input, m.Runes...)
			}
		}
	}

	return n, nil
}

func translateNlCmd(c *client.Client, question, keyspace, table, provider, model string) tea.Cmd {
	return func() tea.Msg {
		if c == nil {
			return nlErrMsg{Err: errNoClient}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		resp, err := c.NlToCql(ctx, question, keyspace, table, provider, model)
		if err != nil {
			return nlErrMsg{Err: err}
		}
		return NlTranslatedMsg{
			CQL:      resp.Cql,
			Provider: resp.Provider,
			Question: question,
			Keyspace: keyspace,
		}
	}
}

func (n NlPanel) View(width int) string {
	if !n.active {
		return ""
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213")).Render("Ask in plain English")
	bodyLines := []string{title}

	if n.loading {
		bodyLines = append(bodyLines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("translating via agent cli..."))
	} else if n.translated != nil {
		amber := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		bodyLines = append(bodyLines, amber.Render("ai-generated via "+n.translated.Provider+" — enter to use, esc to discard"))
		bodyLines = append(bodyLines, lipgloss.NewStyle().Foreground(lipgloss.Color("254")).Render(n.translated.CQL))
	} else {
		if n.err != "" {
			bodyLines = append(bodyLines, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗ "+n.err))
		}
		prompt := "» " + string(n.input)
		bodyLines = append(bodyLines, prompt)

		modelLabel := n.model
		if modelLabel == "" {
			modelLabel = "cli default"
		}
		if n.editingModel {
			bodyLines = append(bodyLines, "model: "+string(n.modelInput))
			bodyLines = append(bodyLines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("enter to save · esc to cancel"))
		} else {
			hint := "enter to translate · ctrl+p provider (" + n.provider + ") · ctrl+e model (" + modelLabel + ") · esc to cancel"
			bodyLines = append(bodyLines, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(hint))
		}
	}

	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("213")).Padding(0, 1)
	content := lipgloss.NewStyle().Width(width - 4).Render(lipgloss.JoinVertical(lipgloss.Left, bodyLines...))
	return box.Render(content)
}
