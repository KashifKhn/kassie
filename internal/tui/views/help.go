package views

import "github.com/charmbracelet/lipgloss"

type HelpView struct {
	Content string
}

func NewHelpView(content string) HelpView {
	return HelpView{Content: content}
}

func (v HelpView) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, v.Content)
}
