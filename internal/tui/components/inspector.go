package components

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

type Inspector struct {
	theme      styles.Theme
	row        *pb.Row
	json       string
	scrollPos  int
	totalLines int
}

func NewInspector(theme styles.Theme) Inspector {
	return Inspector{theme: theme}
}

func (i *Inspector) SetRow(row *pb.Row) {
	i.row = row
	i.json = formatRowJSON(row)
	i.scrollPos = 0
	i.totalLines = strings.Count(i.json, "\n") + 1
}

func (i *Inspector) ScrollDown() {
	if i.scrollPos < i.totalLines-1 {
		i.scrollPos++
	}
}

func (i *Inspector) ScrollUp() {
	if i.scrollPos > 0 {
		i.scrollPos--
	}
}

func (i *Inspector) PageDown(height int) {
	i.scrollPos = minInt(i.scrollPos+height-2, i.totalLines-1)
}

func (i *Inspector) PageUp(height int) {
	i.scrollPos = maxInt(i.scrollPos-height+2, 0)
}

func (i Inspector) CopyToClipboard() error {
	if i.json == "" {
		return fmt.Errorf("nothing to copy")
	}

	return copyToClipboard(i.json)
}

func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else {
			return fmt.Errorf("no clipboard utility found (install xclip, xsel, or wl-copy)")
		}
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("unsupported platform")
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := stdin.Write([]byte(text)); err != nil {
		return err
	}

	stdin.Close()
	return cmd.Wait()
}

func (i Inspector) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	if i.row == nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, i.theme.Dim.Render("Select a row"))
	}

	lines := strings.Split(i.json, "\n")
	i.totalLines = len(lines)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("51")).
		MarginBottom(1)

	header := headerStyle.Render("JSON Inspector")
	contentHeight := height - 3

	if contentHeight < 1 {
		contentHeight = 1
	}

	if i.scrollPos > len(lines)-contentHeight {
		i.scrollPos = maxInt(0, len(lines)-contentHeight)
	}

	endPos := minInt(i.scrollPos+contentHeight, len(lines))
	visibleLines := lines[i.scrollPos:endPos]

	scrollIndicator := ""
	if i.totalLines > contentHeight {
		scrollPercent := float64(i.scrollPos) / float64(i.totalLines-contentHeight) * 100
		scrollIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf(" [%d/%d - %.0f%%]", i.scrollPos+1, i.totalLines, scrollPercent))
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("j/k scroll • d/u page • ctrl+c copy")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header+scrollIndicator,
		"",
		strings.Join(visibleLines, "\n"),
		"",
		footer,
	)

	return lipgloss.NewStyle().Width(width).Height(height).Render(content)
}

func formatRowJSON(row *pb.Row) string {
	if row == nil || row.Cells == nil {
		return ""
	}

	keys := make([]string, 0, len(row.Cells))
	for key := range row.Cells {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	data := make(map[string]any, len(keys))
	for _, key := range keys {
		data[key] = cellToInspectable(row.Cells[key])
	}

	value, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("failed to format row: %v", err)
	}
	return string(value)
}

func cellToInspectable(cell *pb.CellValue) any {
	if cell == nil || cell.IsNull {
		return nil
	}

	switch v := cell.Value.(type) {
	case *pb.CellValue_StringVal:
		return v.StringVal
	case *pb.CellValue_IntVal:
		return v.IntVal
	case *pb.CellValue_DoubleVal:
		return v.DoubleVal
	case *pb.CellValue_BoolVal:
		return v.BoolVal
	case *pb.CellValue_BytesVal:
		return fmt.Sprintf("0x%x", v.BytesVal)
	default:
		return nil
	}
}
