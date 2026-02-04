package db

import (
	"strings"
	"testing"
)

func TestQueryBuilderSelectAll(t *testing.T) {
	tests := []struct {
		name     string
		keyspace string
		table    string
		limit    int
		want     string
	}{
		{
			name:     "without limit",
			keyspace: "mykeyspace",
			table:    "mytable",
			limit:    0,
			want:     "SELECT * FROM mykeyspace.mytable",
		},
		{
			name:     "with limit",
			keyspace: "mykeyspace",
			table:    "mytable",
			limit:    100,
			want:     "SELECT * FROM mykeyspace.mytable LIMIT 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.keyspace, tt.table)
			got := qb.SelectAll(tt.limit)

			if got != tt.want {
				t.Errorf("SelectAll() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryBuilderSelectWithWhere(t *testing.T) {
	tests := []struct {
		name        string
		keyspace    string
		table       string
		whereClause string
		limit       int
		want        string
	}{
		{
			name:        "simple where without limit",
			keyspace:    "mykeyspace",
			table:       "users",
			whereClause: "id = ?",
			limit:       0,
			want:        "SELECT * FROM mykeyspace.users WHERE id = ?",
		},
		{
			name:        "complex where with limit",
			keyspace:    "mykeyspace",
			table:       "users",
			whereClause: "age > ? AND city = ?",
			limit:       50,
			want:        "SELECT * FROM mykeyspace.users WHERE age > ? AND city = ? LIMIT 50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.keyspace, tt.table)
			got := qb.SelectWithWhere(tt.whereClause, tt.limit)

			if got != tt.want {
				t.Errorf("SelectWithWhere() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryBuilderCount(t *testing.T) {
	qb := NewQueryBuilder("mykeyspace", "mytable")
	got := qb.Count()
	want := "SELECT COUNT(*) FROM mykeyspace.mytable"

	if got != want {
		t.Errorf("Count() = %q, want %q", got, want)
	}
}

func TestQueryBuilderInsert(t *testing.T) {
	tests := []struct {
		name     string
		keyspace string
		table    string
		columns  []string
		want     string
	}{
		{
			name:     "single column",
			keyspace: "mykeyspace",
			table:    "users",
			columns:  []string{"id"},
			want:     "INSERT INTO mykeyspace.users (id) VALUES (?)",
		},
		{
			name:     "multiple columns",
			keyspace: "mykeyspace",
			table:    "users",
			columns:  []string{"id", "name", "email"},
			want:     "INSERT INTO mykeyspace.users (id, name, email) VALUES (?, ?, ?)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.keyspace, tt.table)
			got := qb.Insert(tt.columns)

			if got != tt.want {
				t.Errorf("Insert() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryBuilderUpdate(t *testing.T) {
	tests := []struct {
		name        string
		keyspace    string
		table       string
		columns     []string
		whereClause string
		want        string
	}{
		{
			name:        "single column",
			keyspace:    "mykeyspace",
			table:       "users",
			columns:     []string{"name"},
			whereClause: "id = ?",
			want:        "UPDATE mykeyspace.users SET name = ? WHERE id = ?",
		},
		{
			name:        "multiple columns",
			keyspace:    "mykeyspace",
			table:       "users",
			columns:     []string{"name", "email", "age"},
			whereClause: "id = ?",
			want:        "UPDATE mykeyspace.users SET name = ?, email = ?, age = ? WHERE id = ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.keyspace, tt.table)
			got := qb.Update(tt.columns, tt.whereClause)

			if got != tt.want {
				t.Errorf("Update() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryBuilderDelete(t *testing.T) {
	qb := NewQueryBuilder("mykeyspace", "users")
	got := qb.Delete("id = ?")
	want := "DELETE FROM mykeyspace.users WHERE id = ?"

	if got != want {
		t.Errorf("Delete() = %q, want %q", got, want)
	}
}

func TestListKeyspaces(t *testing.T) {
	got := ListKeyspaces()
	want := "SELECT keyspace_name FROM system_schema.keyspaces"

	if got != want {
		t.Errorf("ListKeyspaces() = %q, want %q", got, want)
	}
}

func TestListTables(t *testing.T) {
	got := ListTables("mykeyspace")
	want := "SELECT table_name FROM system_schema.tables WHERE keyspace_name = 'mykeyspace'"

	if got != want {
		t.Errorf("ListTables() = %q, want %q", got, want)
	}
}

func TestDescribeTable(t *testing.T) {
	got := DescribeTable("mykeyspace", "mytable")

	if !strings.Contains(got, "system_schema.columns") {
		t.Error("DescribeTable() should query system_schema.columns")
	}

	if !strings.Contains(got, "mykeyspace") {
		t.Error("DescribeTable() should include keyspace name")
	}

	if !strings.Contains(got, "mytable") {
		t.Error("DescribeTable() should include table name")
	}
}

func TestNewQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder("test_keyspace", "test_table")

	if qb == nil {
		t.Fatal("NewQueryBuilder() returned nil")
	}

	if qb.keyspace != "test_keyspace" {
		t.Errorf("keyspace = %q, want %q", qb.keyspace, "test_keyspace")
	}

	if qb.table != "test_table" {
		t.Errorf("table = %q, want %q", qb.table, "test_table")
	}
}
