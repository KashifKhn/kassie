package service

import (
	"net"
	"testing"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/db"
	"github.com/gocql/gocql"
)

func TestCellToPbValue(t *testing.T) {
	ts := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	uuid, _ := gocql.ParseUUID("550e8400-e29b-41d4-a716-446655440000")
	ip := net.ParseIP("192.168.1.10")

	tests := []struct {
		name      string
		value     interface{}
		cqlType   string
		wantNull  bool
		assertion func(*testing.T, *pb.CellValue)
	}{
		{"nil", nil, "int", true, func(t *testing.T, c *pb.CellValue) {
			if c.CqlType != "int" {
				t.Errorf("nil keeps type label: %q", c.CqlType)
			}
		}},
		{"text", "hello", "text", false, func(t *testing.T, c *pb.CellValue) {
			if c.GetStringVal() != "hello" {
				t.Errorf("string = %q", c.GetStringVal())
			}
		}},
		{"bigint", int64(42), "bigint", false, func(t *testing.T, c *pb.CellValue) {
			if c.GetIntVal() != 42 {
				t.Errorf("int = %d", c.GetIntVal())
			}
		}},
		{"boolean", true, "boolean", false, func(t *testing.T, c *pb.CellValue) {
			if !c.GetBoolVal() {
				t.Error("bool = false")
			}
		}},
		{"double", 3.5, "double", false, func(t *testing.T, c *pb.CellValue) {
			if c.GetDoubleVal() != 3.5 {
				t.Errorf("double = %v", c.GetDoubleVal())
			}
		}},
		{"blob stays raw bytes", []byte{0xDE, 0xAD}, "blob", false, func(t *testing.T, c *pb.CellValue) {
			b := c.GetBytesVal()
			if len(b) != 2 || b[0] != 0xDE {
				t.Errorf("bytes = %x", b)
			}
		}},
		{"timestamp formatted", ts, "timestamp", false, func(t *testing.T, c *pb.CellValue) {
			if c.GetStringVal() != "2026-08-26T12:00:00Z" {
				t.Errorf("timestamp = %q", c.GetStringVal())
			}
		}},
		{"uuid as string", uuid, "uuid", false, func(t *testing.T, c *pb.CellValue) {
			if c.GetStringVal() != "550e8400-e29b-41d4-a716-446655440000" {
				t.Errorf("uuid = %q", c.GetStringVal())
			}
		}},
		{"inet as string", ip, "inet", false, func(t *testing.T, c *pb.CellValue) {
			if c.GetStringVal() != "192.168.1.10" {
				t.Errorf("ip = %q", c.GetStringVal())
			}
		}},
		{
			"map collection to JSON",
			map[string]interface{}{"region": "eu", "tier": int32(3)},
			"frozen<map<text, int>>",
			false,
			func(t *testing.T, c *pb.CellValue) {
				want := `{"region":"eu","tier":3}`
				if c.GetStringVal() != want {
					t.Errorf("map json = %q, want %q", c.GetStringVal(), want)
				}
				if c.CqlType != "frozen<map<text, int>>" {
					t.Errorf("cql type = %q", c.CqlType)
				}
			},
		},
		{
			"list collection to JSON",
			[]interface{}{int64(1), int64(2)},
			"list<bigint>",
			false,
			func(t *testing.T, c *pb.CellValue) {
				if c.GetStringVal() != "[1,2]" {
					t.Errorf("list json = %q", c.GetStringVal())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := cellToPbValue(tt.value, tt.cqlType)
			if cell.IsNull != tt.wantNull {
				t.Fatalf("is_null = %v, want %v", cell.IsNull, tt.wantNull)
			}
			tt.assertion(t, cell)
		})
	}
}

func TestConvertTypedRows(t *testing.T) {
	page := &db.Page{
		Columns: []string{"id", "name"},
		Types:   []string{"int", "text"},
		Rows: [][]interface{}{
			{int32(1), "one"},
			{int32(2), nil},
		},
	}

	rows := convertTypedRows(page)
	if len(rows) != 2 {
		t.Fatalf("rows = %d", len(rows))
	}

	first := rows[0].Cells
	if first["id"].GetIntVal() != 1 || first["id"].CqlType != "int" {
		t.Errorf("id cell = %+v", first["id"])
	}
	if first["name"].GetStringVal() != "one" || first["name"].CqlType != "text" {
		t.Errorf("name cell = %+v", first["name"])
	}

	second := rows[1].Cells
	if !second["name"].IsNull || second["name"].CqlType != "text" {
		t.Errorf("null name cell = %+v", second["name"])
	}
}
