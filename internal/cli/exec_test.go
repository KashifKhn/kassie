package cli

import (
	"strings"
	"testing"
)

func TestSplitStatements(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single", "SELECT * FROM t", []string{"SELECT * FROM t"}},
		{"two with semicolons", "SELECT a FROM x; SELECT b FROM y;", []string{"SELECT a FROM x", "SELECT b FROM y"}},
		{"trailing semicolon optional", "SELECT 1;", []string{"SELECT 1"}},
		{
			"line comments ignored",
			"-- fetch users\nSELECT * FROM users; -- done",
			[]string{"SELECT * FROM users"},
		},
		{
			"block comments ignored",
			"/* header comment */ SELECT 1 /* inline */ ;",
			[]string{"SELECT 1"},
		},
		{
			"semicolons inside strings",
			"SELECT * FROM t WHERE note = 'a;b;c'",
			[]string{"SELECT * FROM t WHERE note = 'a;b;c'"},
		},
		{
			"multiline statement",
			"SELECT id,\n       name\nFROM users\nWHERE id = 1;",
			[]string{"SELECT id,\n       name\nFROM users\nWHERE id = 1"},
		},
		{"empty", "   \n\n  ;\n  -- just a comment", nil},
		{"dashes without comment", "SELECT a-b FROM t", []string{"SELECT a-b FROM t"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitStatements(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d statements, want %d: %q", len(got), len(tt.want), got)
			}
			for i := range got {
				if strings.Join(strings.Fields(got[i]), " ") != strings.Join(strings.Fields(tt.want[i]), " ") {
					t.Errorf("stmt %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSplitStatementsSemicolonInString(t *testing.T) {
	got := splitStatements(`SELECT 'semi;colon' FROM t; SELECT 2 FROM u`)
	if len(got) != 2 {
		t.Fatalf("got %d statements, want 2: %q", len(got), got)
	}
	if !strings.Contains(got[0], "'semi;colon'") {
		t.Errorf("string semicolon mangled: %q", got[0])
	}
}

func TestPreview(t *testing.T) {
	short := preview("SELECT id FROM users")
	if short != "SELECT id FROM users" {
		t.Errorf("preview mangles short stmt: %q", short)
	}

	long := strings.Repeat("x", 100)
	got := preview(long)
	if len(got) > 72 {
		t.Errorf("preview not truncated: %d", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("preview missing ellipsis: %q", got)
	}
}
