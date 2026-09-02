package components

import (
	"bytes"
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
	stats         *pb.TableStats
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

	_ = stdin.Close()
	return cmd.Wait()
}

func (i *Inspector) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	if i.row == nil {
		if i.stats != nil {
			content := lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51")).Render("Table Stats"),
				"",
				formatStatsTable(i.stats),
			)
			return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
		}
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
		if looksLikeCollection(cell.CqlType, v.StringVal) {
			var parsed any
			if err := json.Unmarshal([]byte(v.StringVal), &parsed); err == nil {
				return parsed
			}
		}
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
		key         string
		cqlType     string
		valueLines  []string
		isNull      bool
		isNumber    bool
		isBool      bool
		isString    bool
	}

	rows := make([]rowData, 0, len(keys))
	maxKeyLen := 0
	maxTypeLen := 0

	showTypes := maxWidth >= 60

	for _, key := range keys {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}

		cell := row.Cells[key]
		rd := rowData{key: key, valueLines: []string{"null"}, isNull: true}
		if cell != nil && !cell.IsNull {
			rd.isNull = false
			rd.cqlType = cell.CqlType
			if len(cell.CqlType) > maxTypeLen {
				maxTypeLen = len(cell.CqlType)
			}

			switch v := cell.Value.(type) {
			case *pb.CellValue_StringVal:
				rd.valueLines = collectionValueLines(v.StringVal, cell.CqlType)
				rd.isString = true
			case *pb.CellValue_IntVal:
				rd.valueLines = []string{fmt.Sprintf("%d", v.IntVal)}
				rd.isNumber = true
			case *pb.CellValue_DoubleVal:
				rd.valueLines = []string{fmt.Sprintf("%g", v.DoubleVal)}
				rd.isNumber = true
			case *pb.CellValue_BoolVal:
				rd.valueLines = []string{fmt.Sprintf("%t", v.BoolVal)}
				rd.isBool = true
			case *pb.CellValue_BytesVal:
				rd.valueLines = hexDumpLines(v.BytesVal, 8)
			default:
				rd.valueLines = []string{"null"}
				rd.isNull = true
			}
		}

		rows = append(rows, rd)
	}

	keyColWidth := maxKeyLen
	if maxWidth < 100 {
		keyColWidth = minInt(keyColWidth, 15)
	} else if maxWidth < 150 {
		keyColWidth = minInt(keyColWidth, 20)
	} else {
		keyColWidth = minInt(keyColWidth, 30)
	}

	typeColWidth := 0
	if showTypes && maxTypeLen > 0 {
		typeColWidth = minInt(maxTypeLen, 24)
	}

	valueColWidth := maxWidth - keyColWidth - typeColWidth - 6
	if valueColWidth < 10 {
		valueColWidth = 10
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	nullStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	stringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	boolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	blobStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	var lines []string

	blankKey := padRight("", keyColWidth)
	blankType := ""
	if typeColWidth > 0 {
		blankType = typeStyle.Render(padRight("", typeColWidth) + " │ ")
	}

	for _, rd := range rows {
		keyStr := rd.key
		if len(keyStr) > keyColWidth {
			keyStr = keyStr[:keyColWidth-3] + "..."
		}
		keyCell := keyStyle.Render(padRight(keyStr, keyColWidth))

		typeCell := ""
		if typeColWidth > 0 {
			typeStr := rd.cqlType
			if len(typeStr) > typeColWidth {
				typeStr = typeStr[:typeColWidth-3] + "..."
			}
			typeCell = typeStyle.Render(padRight(typeStr, typeColWidth) + " │ ")
		}

		for i, valueLine := range rd.valueLines {
			if horizontalOffset > 0 && horizontalOffset < len(valueLine) {
				valueLine = valueLine[horizontalOffset:]
			} else if horizontalOffset >= len(valueLine) {
				valueLine = ""
			}
			if len(valueLine) > valueColWidth {
				valueLine = valueLine[:valueColWidth]
			}

			var styleToUse lipgloss.Style
			switch {
			case rd.isNull:
				styleToUse = nullStyle
			case rd.isString:
				styleToUse = stringStyle
			case rd.isNumber:
				styleToUse = numberStyle
			case rd.isBool:
				styleToUse = boolStyle
			default:
				styleToUse = blobStyle
			}

			sep := borderStyle.Render(" │ ")
			if i == 0 {
				lines = append(lines, keyCell+sep+typeCell+styleToUse.Render(valueLine))
			} else {
				lines = append(lines, blankKey+sep+blankType+styleToUse.Render(valueLine))
			}
		}
	}

	return strings.Join(lines, "\n")
}

func collectionValueLines(value string, cqlType string) []string {
	if !looksLikeCollection(cqlType, value) {
		return []string{fmt.Sprintf("%q", value)}
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(value), "", "  "); err != nil {
		return []string{fmt.Sprintf("%q", value)}
	}
	return strings.Split(buf.String(), "\n")
}

func looksLikeCollection(cqlType, value string) bool {
	if strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") {
		return true
	}
	lower := strings.ToLower(cqlType)
	return strings.Contains(lower, "map") || strings.Contains(lower, "list") ||
		strings.Contains(lower, "set") || strings.Contains(lower, "tuple")
}

func hexDumpLines(data []byte, maxLines int) []string {
	if len(data) == 0 {
		return []string{"<empty blob>"}
	}

	var lines []string
	const bytesPerLine = 16

	shown := len(data)
	if maxLines > 0 {
		shown = minInt(shown, maxLines*bytesPerLine)
	}

	for offset := 0; offset < shown; offset += bytesPerLine {
		end := minInt(offset+bytesPerLine, shown)
		chunk := data[offset:end]

		hex := make([]string, len(chunk))
		ascii := make([]rune, len(chunk))
		for i, b := range chunk {
			hex[i] = fmt.Sprintf("%02x", b)
			if b >= 0x20 && b < 0x7f {
				ascii[i] = rune(b)
			} else {
				ascii[i] = '.'
			}
		}

		hexPart := strings.Join(hex, " ")
		if len(chunk) < bytesPerLine {
			hexPart += strings.Repeat("   ", bytesPerLine-len(chunk))
		}

		lines = append(lines, fmt.Sprintf("%08x  %s  |%s|", offset, hexPart, string(ascii)))
	}

	if len(data) > shown {
		lines = append(lines, fmt.Sprintf("... %d more bytes", len(data)-shown))
	}

	return lines
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

func (i *Inspector) SetStats(stats *pb.TableStats) {
	i.stats = stats
}

func (i *Inspector) ClearStats() {
	i.stats = nil
}

func formatStatsTable(stats *pb.TableStats) string {
	if stats == nil {
		return ""
	}

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))

	rows := [][2]string{
		{"rows", formatStatCount(stats.RowCount)},
		{"avg partition", formatStatBytes(stats.MeanPartitionSizeBytes)},
		{"max partition", formatStatBytes(stats.MaxPartitionSizeBytes)},
	}

	var lines []string
	source := "estimate"
	if !stats.EstimateAvailable {
		source = "count(*)"
	}
	lines = append(lines, labelStyle.Render("stats ("+source+")"))
	for _, row := range rows {
		lines = append(lines, "  "+labelStyle.Render(padRight(row[0], 14))+valueStyle.Render(row[1]))
	}
	return strings.Join(lines, "\n")
}

func formatStatCount(n int64) string {
	if n <= 0 {
		return "—"
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func formatStatBytes(n int64) string {
	if n <= 0 {
		return "—"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(n)
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d B", n)
	}
	return fmt.Sprintf("%.1f %s", value, units[unit])
}
