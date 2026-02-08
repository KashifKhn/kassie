package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HelpView struct {
	Content    string
	scrollPos  int
	totalLines int
}

func NewHelpView(content string) HelpView {
	lines := strings.Split(content, "\n")
	return HelpView{
		Content:    content,
		scrollPos:  0,
		totalLines: len(lines),
	}
}

func (v HelpView) Update(msg tea.Msg) (HelpView, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "j", "down":
			if v.scrollPos < v.totalLines {
				v.scrollPos++
			}
		case "k", "up":
			if v.scrollPos > 0 {
				v.scrollPos--
			}
		case "d":
			v.scrollPos += 10
		case "u":
			v.scrollPos -= 10
		case "g":
			v.scrollPos = 0
		case "G":
			v.scrollPos = v.totalLines
		}

		if v.scrollPos < 0 {
			v.scrollPos = 0
		}
		if v.scrollPos > v.totalLines {
			v.scrollPos = v.totalLines
		}
	}
	return v, nil
}

func (v HelpView) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	lines := strings.Split(v.Content, "\n")

	maxScroll := len(lines) - height + 5
	if maxScroll < 0 {
		maxScroll = 0
	}

	scrollPos := v.scrollPos
	if scrollPos > maxScroll {
		scrollPos = maxScroll
	}
	if scrollPos < 0 {
		scrollPos = 0
	}

	visibleHeight := height - 3
	if visibleHeight < 1 {
		visibleHeight = height
	}

	start := scrollPos
	end := start + visibleHeight
	if end > len(lines) {
		end = len(lines)
	}

	visibleLines := lines[start:end]
	content := strings.Join(visibleLines, "\n")

	if len(lines) > visibleHeight {
		scrollPercent := 0
		if maxScroll > 0 {
			scrollPercent = (scrollPos * 100) / maxScroll
			if scrollPercent > 100 {
				scrollPercent = 100
			}
		}

		percentStr := fmt.Sprintf("%d%%", scrollPercent)
		barFilled := scrollPercent / 5
		if barFilled > 20 {
			barFilled = 20
		}
		barEmpty := 20 - barFilled

		footer := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("j/k scroll • d/u page • g/G top/bottom • ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Render(
				"["+strings.Repeat("█", barFilled)+
					strings.Repeat("░", barEmpty)+"] "+
					lipgloss.NewStyle().Bold(true).Render(percentStr))

		content = content + "\n\n" + footer
	}

	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 2)

	return style.Render(content)
}
