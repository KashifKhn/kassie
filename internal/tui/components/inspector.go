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
	theme         styles.Theme
	row           *pb.Row
	json          string
	scrollPos     int
	horizontalPos int
	totalLines    int
	maxLineWidth  int
	displayMode   displayMode
	contentWidth  int
	contentHeight int
	isFullscreen  bool
}

type displayMode int

const (
	displayModeTable displayMode = iota
	displayModePrettyJSON
)

func NewInspector(theme styles.Theme) Inspector {
	return Inspector{
		theme:       theme,
		displayMode: displayModeTable,
	}
}

func (i *Inspector) CycleDisplayMode() {
	i.displayMode = (i.displayMode + 1) % 2
	i.scrollPos = 0
	if i.row != nil && i.contentWidth > 0 {
		i.updateContent()
	}
}

func (i *Inspector) SetRow(row *pb.Row) {
	i.row = row
	i.scrollPos = 0
	if i.contentWidth > 0 {
		i.updateContent()
	}
}

func (i *Inspector) updateContent() {
	switch i.displayMode {
	case displayModeTable:
		i.json = formatRowTable(i.row, i.theme, i.contentWidth, i.horizontalPos)
	case displayModePrettyJSON:
		rawJSON := formatRowJSON(i.row)
		i.json = wrapJSON(rawJSON, i.contentWidth, i.horizontalPos)
	}
	i.totalLines = strings.Count(i.json, "\n") + 1

	i.maxLineWidth = 0
	lines := strings.Split(i.json, "\n")
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth > i.maxLineWidth {
			i.maxLineWidth = lineWidth
		}
	}
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

func (i *Inspector) ScrollLeft() {
	if i.horizontalPos > 0 {
		i.horizontalPos -= 5
		if i.horizontalPos < 0 {
			i.horizontalPos = 0
		}
	}
}

func (i *Inspector) ScrollRight() {
	i.horizontalPos += 5
}

func (i *Inspector) SetFullscreen(fullscreen bool) {
	i.isFullscreen = fullscreen
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

func (i *Inspector) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	if i.row == nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, i.theme.Dim.Render("Select a row"))
	}

	// Store previous state
	prevHorizontal := i.horizontalPos
	prevWidth := i.contentWidth

	// Update dimensions
	i.contentWidth = width
	i.contentHeight = height

	// Only regenerate if something changed
	if i.json == "" || prevWidth != width || prevHorizontal != i.horizontalPos {
		i.updateContent()
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
		modeName = "JSON"
	}

	header := headerStyle.Render(fmt.Sprintf("Inspector [%s]", modeName))

	footerLines := 2

	contentHeight := height - 3 - footerLines

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

	// Add horizontal scroll indicator
	if i.horizontalPos > 0 {
		scrollIndicator += lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf(" [→ %d]", i.horizontalPos))
	}

	var footer string
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	if width < 80 {
		if i.isFullscreen {
			line1 := dimStyle.Render("j/k: scroll • h/l: horizontal")
			line2 := dimStyle.Render("[/]: prev/next row • t: toggle • i: close")
			footer = lipgloss.JoinVertical(lipgloss.Left, line1, line2)
		} else {
			line1 := dimStyle.Render("j/k: scroll • h/l: horizontal")
			line2 := dimStyle.Render("[/]: prev/next row • t: toggle • i: full")
			footer = lipgloss.JoinVertical(lipgloss.Left, line1, line2)
		}
	} else if width < 120 {
		if i.isFullscreen {
			line1 := dimStyle.Render("j/k: scroll • h/l: horiz • [/]: prev/next")
			line2 := dimStyle.Render("t: toggle • i: close • ctrl+c: copy")
			footer = lipgloss.JoinVertical(lipgloss.Left, line1, line2)
		} else {
			line1 := dimStyle.Render("j/k: scroll • h/l: horiz • [/]: prev/next")
			line2 := dimStyle.Render("t: toggle • i: full • ctrl+c: copy")
			footer = lipgloss.JoinVertical(lipgloss.Left, line1, line2)
		}
	} else {
		if i.isFullscreen {
			line1 := dimStyle.Render("j/k: scroll up/down • h/l: scroll left/right • d/u: page down/up")
			line2 := dimStyle.Render("[/]: prev/next row • t: toggle view • i: close fullscreen • ctrl+c: copy")
			footer = lipgloss.JoinVertical(lipgloss.Left, line1, line2)
		} else {
			line1 := dimStyle.Render("j/k: scroll up/down • h/l: scroll left/right • d/u: page down/up")
			line2 := dimStyle.Render("[/]: prev/next row • t: toggle table/json • i: fullscreen • ctrl+c: copy")
			footer = lipgloss.JoinVertical(lipgloss.Left, line1, line2)
		}
	}

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

func formatRowTable(row *pb.Row, theme styles.Theme, maxWidth int, horizontalOffset int) string {
	if row == nil || row.Cells == nil {
		return ""
	}

	keys := make([]string, 0, len(row.Cells))
	for key := range row.Cells {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	type rowData struct {
		key   string
		value string
		raw   any
	}

	rows := make([]rowData, 0, len(keys))
	maxKeyLen := 0

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

		rows = append(rows, rowData{key: key, value: valueStr, raw: value})
	}

	// Fixed key column width - adjust based on available width
	keyColWidth := maxKeyLen

	// In narrow panels, constrain key width more aggressively
	if maxWidth < 100 {
		// Very narrow - limit keys to 15 chars
		if keyColWidth > 15 {
			keyColWidth = 15
		}
	} else if maxWidth < 150 {
		// Narrow - limit keys to 20 chars
		if keyColWidth > 20 {
			keyColWidth = 20
		}
	} else {
		// Wide panel - allow up to 30 chars
		if keyColWidth > 30 {
			keyColWidth = 30
		}
	}

	// Calculate value column width based on remaining space
	valueColWidth := maxWidth - keyColWidth - 4 // 4 for separator and padding
	if valueColWidth < 10 {
		valueColWidth = 10
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

	// No top border - just render rows
	for _, rd := range rows {
		// Truncate key if needed
		keyStr := rd.key
		if len(keyStr) > keyColWidth {
			keyStr = keyStr[:keyColWidth-3] + "..."
		}
		keyPadded := padRight(keyStr, keyColWidth)

		// Apply horizontal scroll to value
		valueStr := rd.value
		if horizontalOffset > 0 && horizontalOffset < len(valueStr) {
			valueStr = valueStr[horizontalOffset:]
		} else if horizontalOffset >= len(valueStr) {
			valueStr = ""
		}

		// Truncate value to visible width (no padding to prevent overflow)
		if len(valueStr) > valueColWidth {
			valueStr = valueStr[:valueColWidth]
		}

		var styleToUse lipgloss.Style
		if rd.raw == nil {
			styleToUse = nullStyle
		} else {
			switch rd.raw.(type) {
			case string:
				styleToUse = stringStyle
			case int64, float64:
				styleToUse = numberStyle
			case bool:
				styleToUse = boolStyle
			default:
				styleToUse = lipgloss.NewStyle()
			}
		}

		separator := borderStyle.Render(" │ ")
		keyCell := keyStyle.Render(keyPadded)
		valueCell := styleToUse.Render(valueStr)

		line := keyCell + separator + valueCell
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}

func wrapJSON(jsonStr string, maxWidth int, horizontalOffset int) string {

	if maxWidth <= 0 || maxWidth < 20 {
		maxWidth = 40
	}

	lines := strings.Split(jsonStr, "\n")
	var processedLines []string

	for _, line := range lines {
		// Apply horizontal scroll
		if horizontalOffset > 0 && horizontalOffset < len(line) {
			line = line[horizontalOffset:]
		} else if horizontalOffset >= len(line) {
			line = ""
		}

		// Truncate to visible width
		runes := []rune(line)
		if len(runes) > maxWidth {
			processedLines = append(processedLines, string(runes[:maxWidth]))
		} else {
			processedLines = append(processedLines, line)
		}
	}

	return strings.Join(processedLines, "\n")
}
