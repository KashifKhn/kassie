package components

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DataGrid struct {
	theme styles.Theme

	keyspace  string
	table     string
	filter    string
	columns   []string
	rows      []rowData
	selected  int
	colOffset int
	cursorID  string
	hasMore   bool
	pageSize  int32
	loading   bool
	status    string
}

type RowSelectedMsg struct {
	Row *pb.Row
}

type dataErrMsg struct {
	Err error
}

type schemaMsg struct {
	Schema *pb.TableSchema
}

type rowsMsg struct {
	Rows     []*pb.Row
	CursorID string
	HasMore  bool
	Filter   string
}

type rowData struct {
	raw  *pb.Row
	cell map[string]string
}

func NewDataGrid(theme styles.Theme) DataGrid {
	return DataGrid{
		theme:    theme,
		pageSize: 50,
		status:   "Select a table",
	}
}

func (g DataGrid) Init() tea.Cmd {
	return nil
}

func (g DataGrid) LoadTable(c *client.Client, keyspace, table string) (DataGrid, tea.Cmd) {
	g.keyspace = keyspace
	g.table = table
	g.filter = ""
	g.columns = nil
	g.rows = nil
	g.selected = 0
	g.colOffset = 0
	g.cursorID = ""
	g.hasMore = false
	g.loading = true
	g.status = fmt.Sprintf("Loading %s.%s...", keyspace, table)

	return g, tea.Batch(
		g.fetchSchemaCmd(c, keyspace, table),
		g.fetchRowsCmd(c, keyspace, table, g.pageSize),
	)
}

func (g DataGrid) ApplyFilter(c *client.Client, where string) (DataGrid, tea.Cmd) {
	if g.keyspace == "" || g.table == "" {
		return g, nil
	}

	g.filter = where
	g.selected = 0
	g.colOffset = 0
	g.cursorID = ""
	g.hasMore = false
	g.loading = true
	g.rows = nil
	if where == "" {
		g.status = "Loading all rows..."
		return g, g.fetchRowsCmd(c, g.keyspace, g.table, g.pageSize)
	}
	g.status = "Filtering..."
	return g, g.fetchFilterCmd(c, g.keyspace, g.table, where, g.pageSize)
}

func (g DataGrid) Refresh(c *client.Client) (DataGrid, tea.Cmd) {
	if g.keyspace == "" || g.table == "" {
		return g, nil
	}
	if g.filter != "" {
		return g.ApplyFilter(c, g.filter)
	}
	return g.LoadTable(c, g.keyspace, g.table)
}

func (g DataGrid) Update(msg tea.Msg, c *client.Client) (DataGrid, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "j", "down":
			if len(g.rows) > 0 {
				g.selected = minInt(g.selected+1, len(g.rows)-1)
			}
		case "k", "up":
			g.selected = maxInt(g.selected-1, 0)
		case "h", "left":
			g.colOffset = maxInt(g.colOffset-1, 0)
		case "l", "right":
			if len(g.columns) > 0 {
				g.colOffset = minInt(g.colOffset+1, len(g.columns)-1)
			}
		case "n":
			if g.hasMore && g.cursorID != "" {
				g.loading = true
				g.status = "Loading next page..."
				return g, g.fetchNextPageCmd(c, g.cursorID, g.filter)
			}
		case "r":
			return g.Refresh(c)
		case "enter":
			if len(g.rows) > 0 {
				row := g.rows[g.selected].raw
				return g, func() tea.Msg { return RowSelectedMsg{Row: row} }
			}
		}
	case schemaMsg:
		g.columns = columnsFromSchema(m.Schema)
	case rowsMsg:
		g.loading = false
		g.cursorID = m.CursorID
		g.hasMore = m.HasMore
		g.filter = m.Filter
		g.rows = convertRows(m.Rows)
		if len(g.rows) == 0 {
			g.status = "No rows"
		} else {
			g.status = fmt.Sprintf("%d rows", len(g.rows))
		}
		g.colOffset = minInt(g.colOffset, maxInt(len(g.columns)-1, 0))
	case dataErrMsg:
		g.loading = false
		g.status = fmt.Sprintf("Error: %s", m.Err)
	}

	return g, nil
}

func (g DataGrid) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	if g.table == "" {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, g.theme.Dim.Render("Select a table"))
	}

	header := g.theme.Accent.Render(fmt.Sprintf("%s.%s", g.keyspace, g.table))
	gridWidth := width
	columns := g.columns
	if len(columns) == 0 && len(g.rows) > 0 {
		columns = columnsFromRows(g.rows)
	}

	colWidths := computeColWidths(columns, g.rows)
	visibleColumns, visibleWidths := fitColumns(columns, colWidths, g.colOffset, gridWidth)

	lines := make([]string, 0, height)
	lines = append(lines, header, "")
	lines = append(lines, renderRow(visibleColumns, visibleWidths, rowData{}, g.theme, false))

	maxRows := height - len(lines) - 1
	if maxRows < 0 {
		maxRows = 0
	}

	for i := 0; i < len(g.rows) && i < maxRows; i++ {
		row := g.rows[i]
		selected := i == g.selected
		lines = append(lines, renderRow(visibleColumns, visibleWidths, row, g.theme, selected))
	}

	footer := g.theme.Status.Render(g.status)
	if g.loading {
		footer = g.theme.Status.Render("Loading...")
	}
	if g.hasMore {
		footer = g.theme.Status.Render(footer + "  (n next)")
	}
	if len(g.columns) > 0 {
		footer = g.theme.Status.Render(footer + "  (h/l scroll, Tab for pane)")
	}
	lines = append(lines, "", footer)

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().Width(width).Height(height).Render(content)
}

func (g DataGrid) Status() string {
	return g.status
}

func (g DataGrid) Keyspace() string {
	return g.keyspace
}

func (g DataGrid) Table() string {
	return g.table
}

func (g DataGrid) Filter() string {
	return g.filter
}

func (g DataGrid) fetchSchemaCmd(c *client.Client, keyspace, table string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		schema, err := c.GetTableSchema(ctx, keyspace, table)
		if err != nil {
			return dataErrMsg{Err: err}
		}
		return schemaMsg{Schema: schema}
	}
}

func (g DataGrid) fetchRowsCmd(c *client.Client, keyspace, table string, pageSize int32) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := c.QueryRows(ctx, keyspace, table, pageSize)
		if err != nil {
			return dataErrMsg{Err: err}
		}
		return rowsMsg{Rows: resp.Rows, CursorID: resp.CursorId, HasMore: resp.HasMore, Filter: ""}
	}
}

func (g DataGrid) fetchNextPageCmd(c *client.Client, cursorID string, filter string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := c.GetNextPage(ctx, cursorID)
		if err != nil {
			return dataErrMsg{Err: err}
		}
		return rowsMsg{Rows: resp.Rows, CursorID: resp.CursorId, HasMore: resp.HasMore, Filter: filter}
	}
}

func (g DataGrid) fetchFilterCmd(c *client.Client, keyspace, table, where string, pageSize int32) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := c.FilterRows(ctx, keyspace, table, where, pageSize)
		if err != nil {
			return dataErrMsg{Err: err}
		}
		return rowsMsg{Rows: resp.Rows, CursorID: resp.CursorId, HasMore: resp.HasMore, Filter: where}
	}
}

func columnsFromSchema(schema *pb.TableSchema) []string {
	if schema == nil {
		return nil
	}
	cols := make([]string, 0, len(schema.Columns))
	for _, col := range schema.Columns {
		cols = append(cols, col.Name)
	}
	return cols
}

func columnsFromRows(rows []rowData) []string {
	set := map[string]struct{}{}
	for _, row := range rows {
		for key := range row.cell {
			set[key] = struct{}{}
		}
	}

	cols := make([]string, 0, len(set))
	for key := range set {
		cols = append(cols, key)
	}
	if len(cols) > 0 {
		sort.Strings(cols)
	}
	return cols
}

func convertRows(rows []*pb.Row) []rowData {
	items := make([]rowData, 0, len(rows))
	for _, row := range rows {
		items = append(items, rowData{
			raw:  row,
			cell: cellMap(row),
		})
	}
	return items
}

func cellMap(row *pb.Row) map[string]string {
	if row == nil || row.Cells == nil {
		return map[string]string{}
	}
	result := make(map[string]string, len(row.Cells))
	for key, cell := range row.Cells {
		result[key] = cellToString(cell)
	}
	return result
}

func cellToString(cell *pb.CellValue) string {
	if cell == nil || cell.IsNull {
		return "null"
	}

	switch v := cell.Value.(type) {
	case *pb.CellValue_StringVal:
		return v.StringVal
	case *pb.CellValue_IntVal:
		return fmt.Sprintf("%d", v.IntVal)
	case *pb.CellValue_DoubleVal:
		return fmt.Sprintf("%g", v.DoubleVal)
	case *pb.CellValue_BoolVal:
		if v.BoolVal {
			return "true"
		}
		return "false"
	case *pb.CellValue_BytesVal:
		return fmt.Sprintf("0x%x", v.BytesVal)
	default:
		return ""
	}
}

func computeColWidths(columns []string, rows []rowData) []int {
	if len(columns) == 0 {
		return nil
	}

	maxWidth := 24
	widths := make([]int, len(columns))

	for i := range widths {
		col := columns[i]
		widths[i] = minInt(maxInt(widths[i], len(col)+2), maxWidth)
		for _, row := range rows {
			value := row.cell[col]
			if value == "" {
				continue
			}
			widths[i] = minInt(maxInt(widths[i], len(value)+2), maxWidth)
		}
	}

	return widths
}

func fitColumns(columns []string, widths []int, offset int, maxWidth int) ([]string, []int) {
	if len(columns) == 0 || len(widths) == 0 || maxWidth <= 0 {
		return nil, nil
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= len(columns) {
		offset = len(columns) - 1
	}

	sepWidth := 3
	used := 0
	cols := make([]string, 0)
	ws := make([]int, 0)

	for i := offset; i < len(columns); i++ {
		needed := widths[i]
		if len(cols) > 0 {
			needed += sepWidth
		}
		if used+needed > maxWidth {
			break
		}
		cols = append(cols, columns[i])
		ws = append(ws, widths[i])
		used += needed
	}

	if len(cols) == 0 {
		cols = append(cols, columns[offset])
		ws = append(ws, minInt(widths[offset], maxWidth))
	}

	return cols, ws
}

func renderRow(columns []string, widths []int, row rowData, theme styles.Theme, selected bool) string {
	parts := make([]string, 0, len(columns))
	for i, col := range columns {
		cell := col
		if row.cell != nil {
			cell = row.cell[col]
			if cell == "" {
				cell = "-"
			}
		}
		cell = truncate(cell, widths[i]-2)
		parts = append(parts, pad(cell, widths[i]))
	}

	line := strings.Join(parts, " │ ")
	if selected {
		return theme.Selected.Render(line)
	}
	if row.cell == nil {
		return theme.Header.Render(line)
	}
	return line
}

func truncate(value string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(value) <= max {
		return value
	}
	if max <= 1 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func pad(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-len(value))
}
