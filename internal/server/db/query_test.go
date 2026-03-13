package db

import (
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
			want:     `SELECT * FROM "mykeyspace"."mytable"`,
		},
		{
			name:     "with limit",
			keyspace: "mykeyspace",
			table:    "mytable",
			limit:    100,
			want:     `SELECT * FROM "mykeyspace"."mytable" LIMIT 100`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb, err := NewQueryBuilder(tt.keyspace, tt.table)
			if err != nil {
				t.Fatalf("NewQueryBuilder() error = %v", err)
			}
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
			want:        `SELECT * FROM "mykeyspace"."users" WHERE id = ?`,
		},
		{
			name:        "complex where with limit",
			keyspace:    "mykeyspace",
			table:       "users",
			whereClause: "age > ? AND city = ?",
			limit:       50,
			want:        `SELECT * FROM "mykeyspace"."users" WHERE age > ? AND city = ? LIMIT 50`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb, err := NewQueryBuilder(tt.keyspace, tt.table)
			if err != nil {
				t.Fatalf("NewQueryBuilder() error = %v", err)
			}
			got := qb.SelectWithWhere(tt.whereClause, tt.limit)

			if got != tt.want {
				t.Errorf("SelectWithWhere() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryBuilderCount(t *testing.T) {
	qb, err := NewQueryBuilder("mykeyspace", "mytable")
	if err != nil {
		t.Fatalf("NewQueryBuilder() error = %v", err)
	}
	got := qb.Count()
	want := `SELECT COUNT(*) FROM "mykeyspace"."mytable"`

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
		wantErr  bool
	}{
		{
			name:     "single column",
			keyspace: "mykeyspace",
			table:    "users",
			columns:  []string{"id"},
			want:     `INSERT INTO "mykeyspace"."users" ("id") VALUES (?)`,
		},
		{
			name:     "multiple columns",
			keyspace: "mykeyspace",
			table:    "users",
			columns:  []string{"id", "name", "email"},
			want:     `INSERT INTO "mykeyspace"."users" ("id", "name", "email") VALUES (?, ?, ?)`,
		},
		{
			name:     "empty columns",
			keyspace: "mykeyspace",
			table:    "users",
			columns:  []string{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb, err := NewQueryBuilder(tt.keyspace, tt.table)
			if err != nil {
				t.Fatalf("NewQueryBuilder() error = %v", err)
			}
			got, err := qb.Insert(tt.columns)
			if tt.wantErr {
				if err == nil {
					t.Error("Insert() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Insert() unexpected error: %v", err)
			}

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
		wantErr     bool
	}{
		{
			name:        "single column",
			keyspace:    "mykeyspace",
			table:       "users",
			columns:     []string{"name"},
			whereClause: "id = ?",
			want:        `UPDATE "mykeyspace"."users" SET "name" = ? WHERE id = ?`,
		},
		{
			name:        "multiple columns",
			keyspace:    "mykeyspace",
			table:       "users",
			columns:     []string{"name", "email", "age"},
			whereClause: "id = ?",
			want:        `UPDATE "mykeyspace"."users" SET "name" = ?, "email" = ?, "age" = ? WHERE id = ?`,
		},
		{
			name:        "empty columns",
			keyspace:    "mykeyspace",
			table:       "users",
			columns:     []string{},
			whereClause: "id = ?",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb, err := NewQueryBuilder(tt.keyspace, tt.table)
			if err != nil {
				t.Fatalf("NewQueryBuilder() error = %v", err)
			}
			got, err := qb.Update(tt.columns, tt.whereClause)
			if tt.wantErr {
				if err == nil {
					t.Error("Update() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Update() unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Update() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryBuilderDelete(t *testing.T) {
	qb, err := NewQueryBuilder("mykeyspace", "users")
	if err != nil {
		t.Fatalf("NewQueryBuilder() error = %v", err)
	}
	got := qb.Delete("id = ?")
	want := `DELETE FROM "mykeyspace"."users" WHERE id = ?`

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
	got, err := ListTables("mykeyspace")
	if err != nil {
		t.Fatalf("ListTables() error = %v", err)
	}
	want := "SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?"

	if got != want {
		t.Errorf("ListTables() = %q, want %q", got, want)
	}
}

func TestListTables_InvalidKeyspace(t *testing.T) {
	_, err := ListTables("bad;keyspace")
	if err == nil {
		t.Error("ListTables() expected error for invalid keyspace, got nil")
	}
}

func TestDescribeTable(t *testing.T) {
	got, err := DescribeTable("mykeyspace", "mytable")
	if err != nil {
		t.Fatalf("DescribeTable() error = %v", err)
	}
	want := "SELECT column_name, type, kind FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ?"

	if got != want {
		t.Errorf("DescribeTable() = %q, want %q", got, want)
	}
}

func TestDescribeTable_InvalidIdentifiers(t *testing.T) {
	tests := []struct {
		name     string
		keyspace string
		table    string
	}{
		{"invalid keyspace", "bad;ks", "table1"},
		{"invalid table", "ks", "bad;table"},
		{"injection attempt", "ks'; DROP KEYSPACE foo; --", "table1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DescribeTable(tt.keyspace, tt.table)
			if err == nil {
				t.Error("DescribeTable() expected error for invalid identifier, got nil")
			}
		})
	}
}

func TestNewQueryBuilder_Validation(t *testing.T) {
	tests := []struct {
		name     string
		keyspace string
		table    string
		wantErr  bool
	}{
		{"valid", "test_keyspace", "test_table", false},
		{"empty keyspace", "", "test_table", true},
		{"empty table", "test_keyspace", "", true},
		{"invalid keyspace chars", "test;keyspace", "test_table", true},
		{"invalid table chars", "test_keyspace", "test;table", true},
		{"injection keyspace", "ks'; DROP TABLE foo; --", "tbl", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb, err := NewQueryBuilder(tt.keyspace, tt.table)
			if tt.wantErr {
				if err == nil {
					t.Error("NewQueryBuilder() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewQueryBuilder() unexpected error: %v", err)
			}
			if qb == nil {
				t.Fatal("NewQueryBuilder() returned nil")
			}
		})
	}
}

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple", "users", `"users"`, false},
		{"with underscore", "my_table", `"my_table"`, false},
		{"empty", "", "", true},
		{"semicolon", "bad;name", "", true},
		{"space", "bad name", "", true},
		{"quote", `bad"name`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QuoteIdentifier(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("QuoteIdentifier() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("QuoteIdentifier() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("QuoteIdentifier() = %q, want %q", got, tt.want)
			}
		})
	}
}
