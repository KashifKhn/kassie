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
	theme       styles.Theme
	row         *pb.Row
	json        string
	scrollPos   int
	totalLines  int
	displayMode displayMode
}

type displayMode int

const (
	displayModeTable displayMode = iota
	displayModePrettyJSON
	displayModeCompactJSON
)

func NewInspector(theme styles.Theme) Inspector {
	return Inspector{
		theme:       theme,
		displayMode: displayModeTable,
	}
}

func (i *Inspector) CycleDisplayMode() {
	i.displayMode = (i.displayMode + 1) % 3
	i.scrollPos = 0
	if i.row != nil {
		i.updateContent()
	}
}

func (i *Inspector) SetRow(row *pb.Row) {
	i.row = row
	i.scrollPos = 0
	i.updateContent()
}

func (i *Inspector) updateContent() {
	switch i.displayMode {
	case displayModeTable:
		i.json = formatRowTable(i.row, i.theme)
	case displayModePrettyJSON:
		i.json = formatRowJSON(i.row)
	case displayModeCompactJSON:
		i.json = formatRowCompact(i.row)
	}
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

	modeName := ""
	switch i.displayMode {
	case displayModeTable:
		modeName = "Table"
	case displayModePrettyJSON:
		modeName = "Pretty JSON"
	case displayModeCompactJSON:
		modeName = "Compact JSON"
	}

	header := headerStyle.Render(fmt.Sprintf("Inspector [%s]", modeName))
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
		Render("j/k scroll • d/u page • t toggle view • ctrl+c copy")

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

func formatRowTable(row *pb.Row, theme styles.Theme) string {
	if row == nil || row.Cells == nil {
		return ""
	}

	keys := make([]string, 0, len(row.Cells))
	for key := range row.Cells {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	maxKeyLen := 0
	maxValueLen := 0

	type rowData struct {
		key   string
		value string
		raw   any
	}

	rows := make([]rowData, 0, len(keys))

	for _, key := range keys {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}

		cell := row.Cells[key]
		value := cellToInspectable(cell)

		var valueStr string
		if value == nil {
			valueStr = "null"
		} else {
			switch v := value.(type) {
			case string:
				valueStr = fmt.Sprintf("\"%s\"", v)
			case int64:
				valueStr = fmt.Sprintf("%d", v)
			case float64:
				valueStr = fmt.Sprintf("%g", v)
			case bool:
				valueStr = fmt.Sprintf("%t", v)
			default:
				valueStr = fmt.Sprintf("%v", v)
			}
		}

		if len(valueStr) > maxValueLen {
			maxValueLen = len(valueStr)
		}

		rows = append(rows, rowData{key: key, value: valueStr, raw: value})
	}

	if maxValueLen > 60 {
		maxValueLen = 60
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("51")).
		Bold(true)

	nullStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	stringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	boolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	var lines []string

	topBorder := "┌" + strings.Repeat("─", maxKeyLen+2) + "┬" + strings.Repeat("─", maxValueLen+2) + "┐"
	lines = append(lines, borderStyle.Render(topBorder))

	for _, rd := range rows {
		keyPadded := padRight(rd.key, maxKeyLen)
		keyRendered := keyStyle.Render(keyPadded)

		valueStr := rd.value
		if len(valueStr) > maxValueLen {
			valueStr = valueStr[:maxValueLen-3] + "..."
		}
		valuePadded := padRight(valueStr, maxValueLen)

		var valueRendered string
		if rd.raw == nil {
			valueRendered = nullStyle.Render(valuePadded)
		} else {
			switch rd.raw.(type) {
			case string:
				valueRendered = stringStyle.Render(valuePadded)
			case int64, float64:
				valueRendered = numberStyle.Render(valuePadded)
			case bool:
				valueRendered = boolStyle.Render(valuePadded)
			default:
				valueRendered = valuePadded
			}
		}

		line := borderStyle.Render("│ ") + keyRendered + borderStyle.Render(" │ ") + valueRendered + borderStyle.Render(" │")
		lines = append(lines, line)
	}

	bottomBorder := "└" + strings.Repeat("─", maxKeyLen+2) + "┴" + strings.Repeat("─", maxValueLen+2) + "┘"
	lines = append(lines, borderStyle.Render(bottomBorder))

	return strings.Join(lines, "\n")
}

func formatRowCompact(row *pb.Row) string {
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

	value, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("failed to format row: %v", err)
	}
	return string(value)
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}
