package styles

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Title      lipgloss.Style
	Border     lipgloss.Style
	BorderHot  lipgloss.Style
	Dim        lipgloss.Style
	Error      lipgloss.Style
	Accent     lipgloss.Style
	Status     lipgloss.Style
	Background lipgloss.Style
	Panel      lipgloss.Style
	PanelHot   lipgloss.Style
	Header     lipgloss.Style
	SidebarKey lipgloss.Style
	SidebarTbl lipgloss.Style
	Selected   lipgloss.Style
}

func DefaultTheme() Theme {
	base := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	panel := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("238"))
	panelHot := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("69"))

	return Theme{
		Title:      base.Bold(true),
		Border:     base.Foreground(lipgloss.Color("238")),
		BorderHot:  base.Foreground(lipgloss.Color("69")),
		Dim:        base.Foreground(lipgloss.Color("242")),
		Error:      base.Foreground(lipgloss.Color("196")),
		Accent:     base.Foreground(lipgloss.Color("33")).Bold(true),
		Status:     base.Foreground(lipgloss.Color("108")),
		Background: base,
		Panel:      panel,
		PanelHot:   panelHot,
		Header:     base.Foreground(lipgloss.Color("110")).Bold(true),
		SidebarKey: base.Foreground(lipgloss.Color("110")),
		SidebarTbl: base.Foreground(lipgloss.Color("250")),
		Selected:   base.Background(lipgloss.Color("24")).Foreground(lipgloss.Color("255")).Bold(true),
	}
}
