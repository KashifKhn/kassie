package service

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/gocql/gocql"
)

func csvLines(t *testing.T, data []byte) [][]string {
	t.Helper()
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	return records
}

func TestCSVHeader(t *testing.T) {
	data, err := csvHeader([]string{"id", "name", "active"})
	if err != nil {
		t.Fatalf("csvHeader: %v", err)
	}
	lines := csvLines(t, data)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0][0] != "id" || lines[0][1] != "name" || lines[0][2] != "active" {
		t.Errorf("unexpected header: %v", lines[0])
	}
}

func TestCSVHeaderEscapesCommas(t *testing.T) {
	data, err := csvHeader([]string{"a,b", "c"})
	if err != nil {
		t.Fatalf("csvHeader: %v", err)
	}
	lines := csvLines(t, data)
	if len(lines[0]) != 2 || lines[0][0] != "a,b" {
		t.Errorf("comma in column name not escaped: %q", data)
	}
}

func TestCSVRow(t *testing.T) {
	ts := time.Date(2026, 8, 26, 10, 30, 0, 0, time.UTC)
	uuid, err0 := gocql.ParseUUID("550e8400-e29b-41d4-a716-446655440000")
	if err0 != nil {
		t.Fatalf("parse uuid: %v", err0)
	}

	tests := []struct {
		name string
		row  []interface{}
		want []string
	}{
		{"primitives", []interface{}{int64(1), "text", true}, []string{"1", "text", "true"}},
		{"nil becomes empty", []interface{}{nil, int64(2)}, []string{"", "2"}},
		{"timestamp RFC3339", []interface{}{ts}, []string{"2026-08-26T10:30:00Z"}},
		{"uuid", []interface{}{uuid}, []string{"550e8400-e29b-41d4-a716-446655440000"}},
		{"blob as text", []interface{}{[]byte("raw")}, []string{"raw"}},
		{
			"quotes escaped",
			[]interface{}{`say "hi"`},
			[]string{`say "hi"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := csvRow(nil, tt.row)
			if err != nil {
				t.Fatalf("csvRow: %v", err)
			}
			got := csvLines(t, data)
			if len(got) != 1 {
				t.Fatalf("expected 1 record, got %d", len(got))
			}
			for i, want := range tt.want {
				if got[0][i] != want {
					t.Errorf("field %d = %q, want %q", i, got[0][i], want)
				}
			}
		})
	}
}

func TestJSONRow(t *testing.T) {
	row := []interface{}{int32(7), "hello", nil, 3.5, false}
	data, err := jsonRow([]string{"id", "label", "missing", "score", "flag"}, row)
	if err != nil {
		t.Fatalf("jsonRow: %v", err)
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, data)
	}

	if obj["id"] != float64(7) {
		t.Errorf("id = %v, want 7", obj["id"])
	}
	if obj["label"] != "hello" {
		t.Errorf("label = %v", obj["label"])
	}
	if v, ok := obj["missing"]; !ok || v != nil {
		t.Errorf("missing = %v, want present null", v)
	}
	if obj["score"] != 3.5 || obj["flag"] != false {
		t.Errorf("unexpected values: %v", obj)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Error("JSON rows must be newline-delimited")
	}
}

func TestJSONRowUUID(t *testing.T) {
	uuid, err0 := gocql.ParseUUID("550e8400-e29b-41d4-a716-446655440000")
	if err0 != nil {
		t.Fatalf("parse uuid: %v", err0)
	}
	data, err := jsonRow([]string{"id"}, []interface{}{uuid})
	if err != nil {
		t.Fatalf("jsonRow: %v", err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if obj["id"] != uuid.String() {
		t.Errorf("uuid serialized as %v, want string form", obj["id"])
	}
}

func TestFormatterFor(t *testing.T) {
	tests := []struct {
		format pb.ExportFormat
		valid  bool
	}{
		{pb.ExportFormat_EXPORT_FORMAT_CSV, true},
		{pb.ExportFormat_EXPORT_FORMAT_JSON, true},
		{pb.ExportFormat_EXPORT_FORMAT_UNSPECIFIED, false},
		{pb.ExportFormat(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.format.String(), func(t *testing.T) {
			_, _, err := formatterFor(tt.format)
			if tt.valid && err != nil {
				t.Errorf("format %v: unexpected error %v", tt.format, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("format %v: expected error", tt.format)
			}
		})
	}
}
