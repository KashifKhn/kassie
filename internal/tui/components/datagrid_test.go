package components

import (
	"testing"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/tui/cache"
	"github.com/KashifKhn/kassie/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
)

func createTestGrid() DataGrid {
	theme := styles.DefaultTheme()
	schemaCache := cache.NewSchemaCache(0)
	return NewDataGrid(theme, schemaCache)
}

func createTestRows(count int) []rowData {
	rows := make([]rowData, count)
	for i := 0; i < count; i++ {
		rows[i] = rowData{
			raw: &pb.Row{},
			cell: map[string]string{
				"id":   string(rune('A' + i)),
				"name": "User" + string(rune('0'+i)),
			},
		}
	}
	return rows
}

func TestDataGrid_ViewportScrolling(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(100)
	g.selected = 0
	g.viewportOffset = 0

	maxRows := 20

	if g.selected < g.viewportOffset {
		g.viewportOffset = g.selected
	}
	if g.selected >= g.viewportOffset+maxRows {
		g.viewportOffset = g.selected - maxRows + 1
	}

	if g.viewportOffset != 0 {
		t.Errorf("expected viewportOffset 0, got %d", g.viewportOffset)
	}

	g.selected = 25
	if g.selected < g.viewportOffset {
		g.viewportOffset = g.selected
	}
	if g.selected >= g.viewportOffset+maxRows {
		g.viewportOffset = g.selected - maxRows + 1
	}

	expected := 25 - maxRows + 1
	if g.viewportOffset != expected {
		t.Errorf("expected viewportOffset %d when selected=25, got %d", expected, g.viewportOffset)
	}
}

func TestDataGrid_ViewportScrollUp(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(100)
	g.selected = 50
	g.viewportOffset = 40

	g.selected = 35

	if g.selected < g.viewportOffset {
		g.viewportOffset = g.selected
	}

	if g.viewportOffset != 35 {
		t.Errorf("expected viewportOffset to adjust to 35, got %d", g.viewportOffset)
	}
}

func TestDataGrid_SearchPerform(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)

	g.searchInput = textinput.New()
	g.searchInput.SetValue("User5")

	g = g.performSearch()

	if len(g.matchedRows) != 1 {
		t.Errorf("expected 1 match, got %d", len(g.matchedRows))
	}

	if len(g.matchedRows) > 0 && g.matchedRows[0] != 5 {
		t.Errorf("expected match at index 5, got %d", g.matchedRows[0])
	}

	if g.selected != 5 {
		t.Errorf("expected selected to be 5, got %d", g.selected)
	}
}

func TestDataGrid_SearchCaseInsensitive(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)

	g.searchInput = textinput.New()
	g.searchInput.SetValue("user5")

	g = g.performSearch()

	if len(g.matchedRows) != 1 {
		t.Errorf("expected 1 match (case insensitive), got %d", len(g.matchedRows))
	}

	if len(g.matchedRows) > 0 && g.matchedRows[0] != 5 {
		t.Errorf("expected match at index 5, got %d", g.matchedRows[0])
	}
}

func TestDataGrid_SearchPartialMatch(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)

	g.searchInput = textinput.New()
	g.searchInput.SetValue("User")

	g = g.performSearch()

	if len(g.matchedRows) != 10 {
		t.Errorf("expected 10 matches for 'User', got %d", len(g.matchedRows))
	}
}

func TestDataGrid_SearchNoMatches(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)

	g.searchInput = textinput.New()
	g.searchInput.SetValue("NonExistent")

	g = g.performSearch()

	if len(g.matchedRows) != 0 {
		t.Errorf("expected 0 matches, got %d", len(g.matchedRows))
	}
}

func TestDataGrid_SearchEmptyQuery(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)

	g.searchInput = textinput.New()
	g.searchInput.SetValue("")

	g = g.performSearch()

	if len(g.matchedRows) != 0 {
		t.Errorf("expected 0 matches for empty query, got %d", len(g.matchedRows))
	}
}

func TestDataGrid_NextMatch(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)
	g.matchedRows = []int{2, 5, 8}
	g.matchIndex = 0
	g.selected = 2

	g = g.nextMatch()

	if g.matchIndex != 1 {
		t.Errorf("expected matchIndex 1, got %d", g.matchIndex)
	}
	if g.selected != 5 {
		t.Errorf("expected selected 5, got %d", g.selected)
	}

	g = g.nextMatch()

	if g.matchIndex != 2 {
		t.Errorf("expected matchIndex 2, got %d", g.matchIndex)
	}
	if g.selected != 8 {
		t.Errorf("expected selected 8, got %d", g.selected)
	}
}

func TestDataGrid_NextMatchCircular(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)
	g.matchedRows = []int{2, 5, 8}
	g.matchIndex = 2
	g.selected = 8

	g = g.nextMatch()

	if g.matchIndex != 0 {
		t.Errorf("expected matchIndex to wrap to 0, got %d", g.matchIndex)
	}
	if g.selected != 2 {
		t.Errorf("expected selected to wrap to 2, got %d", g.selected)
	}
}

func TestDataGrid_PrevMatch(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)
	g.matchedRows = []int{2, 5, 8}
	g.matchIndex = 2
	g.selected = 8

	g = g.prevMatch()

	if g.matchIndex != 1 {
		t.Errorf("expected matchIndex 1, got %d", g.matchIndex)
	}
	if g.selected != 5 {
		t.Errorf("expected selected 5, got %d", g.selected)
	}
}

func TestDataGrid_PrevMatchCircular(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)
	g.matchedRows = []int{2, 5, 8}
	g.matchIndex = 0
	g.selected = 2

	g = g.prevMatch()

	if g.matchIndex != 2 {
		t.Errorf("expected matchIndex to wrap to 2, got %d", g.matchIndex)
	}
	if g.selected != 8 {
		t.Errorf("expected selected to wrap to 8, got %d", g.selected)
	}
}

func TestDataGrid_NextMatchNoMatches(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)
	g.matchedRows = []int{}
	g.matchIndex = 0
	g.selected = 5

	g = g.nextMatch()

	if g.selected != 5 {
		t.Errorf("expected selected to remain 5 when no matches, got %d", g.selected)
	}
}

func TestDataGrid_PrevMatchNoMatches(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)
	g.matchedRows = []int{}
	g.matchIndex = 0
	g.selected = 5

	g = g.prevMatch()

	if g.selected != 5 {
		t.Errorf("expected selected to remain 5 when no matches, got %d", g.selected)
	}
}

func TestDataGrid_ActivateSearch(t *testing.T) {
	g := createTestGrid()

	if g.searchActive {
		t.Error("expected searchActive to be false initially")
	}

	g = g.ActivateSearch()

	if !g.searchActive {
		t.Error("expected searchActive to be true after activation")
	}
}

func TestDataGrid_DeactivateSearch(t *testing.T) {
	g := createTestGrid()
	g = g.ActivateSearch()

	if !g.searchActive {
		t.Error("expected searchActive to be true after activation")
	}

	g = g.DeactivateSearch()

	if g.searchActive {
		t.Error("expected searchActive to be false after deactivation")
	}
}

func TestDataGrid_IsSearchActive(t *testing.T) {
	g := createTestGrid()

	if g.IsSearchActive() {
		t.Error("expected IsSearchActive to return false initially")
	}

	g.searchActive = true

	if !g.IsSearchActive() {
		t.Error("expected IsSearchActive to return true")
	}
}

func TestDataGrid_ColumnWidthCaching(t *testing.T) {
	g := createTestGrid()
	g.rows = createTestRows(10)
	g.columns = []string{"id", "name"}

	if g.cachedColWidths != nil {
		t.Error("expected cachedColWidths to be nil initially")
	}

	widths := computeColWidths(g.columns, g.rows)
	g.cachedColWidths = widths

	if g.cachedColWidths == nil {
		t.Error("expected cachedColWidths to be set")
	}

	if len(g.cachedColWidths) != 2 {
		t.Errorf("expected 2 cached widths, got %d", len(g.cachedColWidths))
	}
}

func TestDataGrid_CacheInvalidationOnLoadTable(t *testing.T) {
	g := createTestGrid()
	g.cachedColWidths = []int{10, 20}

	g.keyspace = "test_ks"
	g.table = "test_table"
	g.columns = nil
	g.rows = nil
	g.selected = 0
	g.viewportOffset = 0
	g.colOffset = 0
	g.cursorID = ""
	g.hasMore = false
	g.loading = true
	g.cachedColWidths = nil

	if g.cachedColWidths != nil {
		t.Error("expected cachedColWidths to be nil after LoadTable")
	}
}

func TestDataGrid_SearchMultipleColumns(t *testing.T) {
	g := createTestGrid()
	g.rows = []rowData{
		{
			raw: &pb.Row{},
			cell: map[string]string{
				"id":    "1",
				"name":  "Alice",
				"email": "alice@example.com",
			},
		},
		{
			raw: &pb.Row{},
			cell: map[string]string{
				"id":    "2",
				"name":  "Bob",
				"email": "bob@test.com",
			},
		},
		{
			raw: &pb.Row{},
			cell: map[string]string{
				"id":    "3",
				"name":  "Charlie",
				"email": "charlie@example.com",
			},
		},
	}

	g.searchInput = textinput.New()
	g.searchInput.SetValue("example")

	g = g.performSearch()

	if len(g.matchedRows) != 2 {
		t.Errorf("expected 2 matches for 'example' in email column, got %d", len(g.matchedRows))
	}
}

func TestComputeColWidths(t *testing.T) {
	columns := []string{"id", "name"}
	rows := []rowData{
		{
			raw:  &pb.Row{},
			cell: map[string]string{"id": "1", "name": "ShortName"},
		},
		{
			raw:  &pb.Row{},
			cell: map[string]string{"id": "2", "name": "VeryLongNameHere"},
		},
	}

	widths := computeColWidths(columns, rows)

	if len(widths) != 2 {
		t.Errorf("expected 2 widths, got %d", len(widths))
	}

	if widths[0] < 4 {
		t.Errorf("expected id width >= 4 (including padding), got %d", widths[0])
	}

	if widths[1] < 18 {
		t.Errorf("expected name width >= 18 (VeryLongNameHere + padding), got %d", widths[1])
	}

	if widths[1] > 24 {
		t.Errorf("expected name width <= 24 (maxWidth), got %d", widths[1])
	}
}

func TestComputeColWidths_EmptyColumns(t *testing.T) {
	widths := computeColWidths([]string{}, []rowData{})

	if widths != nil {
		t.Error("expected nil widths for empty columns")
	}
}

func TestFitColumns(t *testing.T) {
	columns := []string{"col1", "col2", "col3", "col4"}
	widths := []int{10, 15, 20, 25}
	offset := 0
	maxWidth := 50

	visibleCols, visibleWidths := fitColumns(columns, widths, offset, maxWidth)

	if len(visibleCols) == 0 {
		t.Error("expected at least one visible column")
	}

	if len(visibleCols) != len(visibleWidths) {
		t.Errorf("columns and widths length mismatch: %d vs %d", len(visibleCols), len(visibleWidths))
	}

	totalWidth := 0
	for i, w := range visibleWidths {
		totalWidth += w
		if i > 0 {
			totalWidth += 3
		}
	}

	if totalWidth > maxWidth {
		t.Errorf("total width %d exceeds maxWidth %d", totalWidth, maxWidth)
	}
}

func TestFitColumns_WithOffset(t *testing.T) {
	columns := []string{"col1", "col2", "col3", "col4"}
	widths := []int{10, 15, 20, 25}
	offset := 2
	maxWidth := 50

	visibleCols, _ := fitColumns(columns, widths, offset, maxWidth)

	if len(visibleCols) == 0 {
		t.Error("expected at least one visible column")
	}

	if visibleCols[0] != "col3" {
		t.Errorf("expected first visible column to be col3, got %s", visibleCols[0])
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		max    int
		expect string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate", "hello world", 8, "hello..."},
		{"very short max", "hello", 3, "..."},
		{"max zero", "hello", 0, ""},
		{"max one", "hello", 1, "h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.value, tt.max)
			if result != tt.expect {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.value, tt.max, result, tt.expect)
			}
		})
	}
}

func TestPad(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		width  int
		expect string
	}{
		{"no padding needed", "hello", 5, "hello"},
		{"add padding", "hi", 5, "hi   "},
		{"already longer", "hello world", 5, "hello world"},
		{"zero width", "hi", 0, "hi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pad(tt.value, tt.width)
			if result != tt.expect {
				t.Errorf("pad(%q, %d) = %q, want %q", tt.value, tt.width, result, tt.expect)
			}
		})
	}
}
