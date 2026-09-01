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
