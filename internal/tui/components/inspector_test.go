package components

import (
	"strings"
	"testing"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/tui/styles"
)

func TestHexDumpLines(t *testing.T) {
	data := make([]byte, 20)
	for i := range data {
		data[i] = byte(0x20 + i)
	}

	lines := hexDumpLines(data, 8)
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2 (20 bytes = 2 rows of 16)", len(lines))
	}
	if !strings.HasPrefix(lines[0], "00000000  20 21 22 23") {
		t.Errorf("first line = %q", lines[0])
	}
	if !strings.Contains(lines[0], "| !\"#") {
		t.Errorf("ascii column missing: %q", lines[0])
	}
	if !strings.Contains(lines[1], "00000010") {
		t.Errorf("offset wrong: %q", lines[1])
	}
}

func TestHexDumpLinesTruncates(t *testing.T) {
	data := make([]byte, 100)
	lines := hexDumpLines(data, 4)
	if len(lines) != 5 {
		t.Fatalf("lines = %d, want 5 (4 rows + overflow notice)", len(lines))
	}
	if !strings.Contains(lines[4], "36 more bytes") {
		t.Errorf("overflow notice = %q", lines[4])
	}
}

func TestHexDumpLinesEmpty(t *testing.T) {
	if got := hexDumpLines(nil, 8); got[0] != "<empty blob>" {
		t.Errorf("empty blob = %q", got[0])
	}
}

func TestCollectionValueLinesJSON(t *testing.T) {
	lines := collectionValueLines(`{"a":1,"b":[2,3]}`, "map<varchar, int>")
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "\"a\": 1") || !strings.Contains(joined, "\"b\": [") {
		t.Errorf("pretty JSON expected, got: %q", joined)
	}

	single := collectionValueLines(`just text`, "text")
	if len(single) != 1 || single[0] != `"just text"` {
		t.Errorf("plain string = %v", single)
	}
}

func TestLooksLikeCollection(t *testing.T) {
	if !looksLikeCollection("map<text,int>", "{}") {
		t.Error("map type should be collection")
	}
	if !looksLikeCollection("text", "[1,2]") {
		t.Error("json array value should be collection")
	}
	if looksLikeCollection("text", "hello") {
		t.Error("plain string should not be collection")
	}
}

func TestFormatRowTableShowsTypes(t *testing.T) {
	theme := styles.DefaultTheme()
	row := &pb.Row{Cells: map[string]*pb.CellValue{
		"id": {Value: &pb.CellValue_IntVal{IntVal: 7}, CqlType: "int"},
		"blobcol": {Value: &pb.CellValue_BytesVal{BytesVal: []byte{0xDE, 0xAD}}, CqlType: "blob"},
	}}

	out := formatRowTable(row, theme, 120, 0)

	if !strings.Contains(out, "int") {
		t.Error("type column missing int")
	}
	if !strings.Contains(out, "blob") {
		t.Error("type column missing blob")
	}
	if !strings.Contains(out, "de ad") {
		t.Error("hex dump missing")
	}
}

func TestFormatRowTableNullCell(t *testing.T) {
	theme := styles.DefaultTheme()
	row := &pb.Row{Cells: map[string]*pb.CellValue{
		"missing": {IsNull: true, CqlType: "text"},
	}}

	out := formatRowTable(row, theme, 120, 0)
	if !strings.Contains(out, "null") {
		t.Error("null value not rendered")
	}
}

func TestFormatRowJSONUnwrapsCollections(t *testing.T) {
	row := &pb.Row{Cells: map[string]*pb.CellValue{
		"attrs": {
			Value:   &pb.CellValue_StringVal{StringVal: `{"a":1}`},
			CqlType: "map<varchar, int>",
		},
	}}

	out := formatRowJSON(row)
	if !strings.Contains(out, `"a": 1`) {
		t.Errorf("collection not unwrapped in JSON mode: %q", out)
	}
}

func TestFormatStatsTable(t *testing.T) {
	out := formatStatsTable(&pb.TableStats{
		RowCount:                1234567,
		MeanPartitionSizeBytes:  2048,
		MaxPartitionSizeBytes:   4096,
		EstimateAvailable:       true,
	})

	if !strings.Contains(out, "1.2M") {
		t.Errorf("row count formatting: %q", out)
	}
	if !strings.Contains(out, "2.0 KB") {
		t.Errorf("bytes formatting: %q", out)
	}
	if !strings.Contains(out, "estimate") {
		t.Errorf("source label: %q", out)
	}
}

func TestFormatStatsCountFallback(t *testing.T) {
	out := formatStatsTable(&pb.TableStats{RowCount: 42, EstimateAvailable: false})
	if !strings.Contains(out, "count(*)") {
		t.Errorf("fallback label: %q", out)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("count: %q", out)
	}
}

func TestInspectorShowsStatsWithoutRow(t *testing.T) {
	insp := NewInspector(styles.DefaultTheme())
	insp.SetStats(&pb.TableStats{RowCount: 5})

	view := insp.View(80, 24)
	if !strings.Contains(view, "Table Stats") {
		t.Errorf("stats not rendered when no row selected: %q", view)
	}
}

func TestFormatTraceWaterfall(t *testing.T) {
	trace := &pb.GetTraceResponse{
		Ready:       true,
		DurationUs:  10000,
		Coordinator: "127.0.0.1",
		Events: []*pb.TraceEvent{
			{Activity: "Parsing statement", Source: "127.0.0.1", ElapsedUs: 100},
			{Activity: "Executing single-partition query on users", Source: "127.0.0.1", ElapsedUs: 9000},
		},
	}

	out := formatTrace(trace, 100)
	if !strings.Contains(out, "total") {
		t.Errorf("total line missing: %q", out)
	}
	if !strings.Contains(out, "9.0ms") {
		t.Errorf("event elapsed formatting missing: %q", out)
	}
	if !strings.Contains(out, "█") {
		t.Errorf("waterfall bars missing: %q", out)
	}
	if !strings.Contains(out, "Parsing statement") {
		t.Errorf("activity missing: %q", out)
	}
}

func TestFormatTraceNotReady(t *testing.T) {
	out := formatTrace(&pb.GetTraceResponse{Ready: false}, 100)
	if !strings.Contains(out, "not ready") {
		t.Errorf("not-ready hint missing: %q", out)
	}
}

func TestFormatMicros(t *testing.T) {
	tests := []struct {
		us   int64
		want string
	}{
		{500, "500µs"},
		{1500, "1.5ms"},
		{2500000, "2.50s"},
		{-1, "?"},
	}
	for _, tt := range tests {
		if got := formatMicros(tt.us); got != tt.want {
			t.Errorf("formatMicros(%d) = %q, want %q", tt.us, got, tt.want)
		}
	}
}

func TestInspectorTraceMode(t *testing.T) {
	insp := NewInspector(styles.DefaultTheme())
	insp.SetTrace(&pb.GetTraceResponse{
		Ready:      true,
		DurationUs: 5000,
		Coordinator: "node1",
		Events: []*pb.TraceEvent{{Activity: "exec", ElapsedUs: 4000}},
	})

	view := insp.View(80, 24)
	if !strings.Contains(view, "Query Trace") {
		t.Errorf("trace header missing: %q", view)
	}
	if !strings.Contains(view, "█") {
		t.Errorf("waterfall bars missing: %q", view)
	}
	if !insp.HasTrace() {
		t.Error("HasTrace false after SetTrace")
	}
}
