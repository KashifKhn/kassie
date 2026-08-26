package service

import (
	"strings"
	"testing"
)

func TestNormalizeAdhocCQL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{
			name:  "plain select",
			input: "SELECT * FROM ks.tbl",
			want:  "SELECT * FROM ks.tbl",
		},
		{
			name:  "case insensitive prefix",
			input: "select id from t",
			want:  "select id from t",
		},
		{
			name:  "leading whitespace and trailing semicolon",
			input: "  \n\tSELECT id FROM t ;\n ",
			want:  "SELECT id FROM t",
		},
		{
			name:  "single trailing semicolon only is stripped",
			input: "SELECT a FROM b;",
			want:  "SELECT a FROM b",
		},
		{
			name:    "multiple statements rejected",
			input:   "SELECT 1; DROP TABLE x",
			wantErr: "multiple statements",
		},
		{
			name:    "empty rejected",
			input:   "   ",
			wantErr: "required",
		},
		{
			name:    "non-select rejected",
			input:   "INSERT INTO t (a) VALUES (1)",
			wantErr: "only SELECT statements",
		},
		{
			name:    "delete rejected",
			input:   "DELETE FROM t WHERE a = 1",
			wantErr: "only SELECT statements",
		},
		{
			name:    "update keyword hidden in select rejected",
			input:   "SELECT * FROM t ; UPDATE x",
			wantErr: "multiple statements",
		},
		{
			name:    "use keyword inside select rejected",
			input:   "SELECT * FROM use",
			wantErr: "disallowed keywords",
		},
		{
			name:    "drop as identifier rejected",
			input:   "SELECT drop FROM t",
			wantErr: "disallowed keywords",
		},
		{
			name:    "begin batch smuggled in function call rejected",
			input:   "SELECT f(BEGIN BATCH) FROM t",
			wantErr: "disallowed keywords",
		},
		{
			name:  "keywords embedded in identifiers allowed when not standalone",
			input: "SELECT used_count, insertions FROM metrics_daily",
			want:  "SELECT used_count, insertions FROM metrics_daily",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAdhocCQL(tt.input)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (result %q)", tt.wantErr, got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("normalized = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeAdhocCQL_MaxLength(t *testing.T) {
	long := "SELECT " + strings.Repeat("a", 10*1024)
	if _, err := normalizeAdhocCQL(long); err == nil {
		t.Fatal("expected max length error")
	}
}
