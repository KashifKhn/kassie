package components

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

type Inspector struct {
	theme styles.Theme
	row   *pb.Row
	json  string
}

func NewInspector(theme styles.Theme) Inspector {
	return Inspector{theme: theme}
}

func (i *Inspector) SetRow(row *pb.Row) {
	i.row = row
	i.json = formatRowJSON(row)
}

func (i Inspector) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	if i.row == nil {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, i.theme.Dim.Render("Select a row"))
	}

	lines := strings.Split(i.json, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}

	content := strings.Join(lines, "\n")
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
