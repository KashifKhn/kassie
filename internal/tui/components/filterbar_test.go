package components

import (
	"testing"
)

func TestValidateFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		errMatch string
	}{
		{
			name:    "empty filter is valid",
			input:   "",
			wantErr: false,
		},
		{
			name:    "whitespace only is valid",
			input:   "   ",
			wantErr: false,
		},
		{
			name:    "simple equality",
			input:   "user_id = 123",
			wantErr: false,
		},
		{
			name:    "greater than",
			input:   "age > 18",
			wantErr: false,
		},
		{
			name:    "less than or equal",
			input:   "score <= 100",
			wantErr: false,
		},
		{
			name:    "IN clause",
			input:   "status IN ('active', 'pending')",
			wantErr: false,
		},
		{
			name:    "CONTAINS keyword",
			input:   "tags CONTAINS 'urgent'",
			wantErr: false,
		},
		{
			name:    "AND operator",
			input:   "age > 18 AND status = 'active'",
			wantErr: false,
		},
		{
			name:    "OR operator",
			input:   "role = 'admin' OR role = 'moderator'",
			wantErr: false,
		},
		{
			name:     "dangerous DROP keyword",
			input:    "DROP TABLE users",
			wantErr:  true,
			errMatch: "DROP",
		},
		{
			name:     "dangerous DELETE keyword",
			input:    "DELETE FROM users",
			wantErr:  true,
			errMatch: "DELETE",
		},
		{
			name:     "dangerous TRUNCATE keyword",
			input:    "TRUNCATE users",
			wantErr:  true,
			errMatch: "TRUNCATE",
		},
		{
			name:     "dangerous ALTER keyword",
			input:    "ALTER TABLE users",
			wantErr:  true,
			errMatch: "ALTER",
		},
		{
			name:     "dangerous CREATE keyword",
			input:    "CREATE TABLE evil",
			wantErr:  true,
			errMatch: "CREATE",
		},
		{
			name:     "dangerous INSERT keyword",
			input:    "INSERT INTO users",
			wantErr:  true,
			errMatch: "INSERT",
		},
		{
			name:     "dangerous UPDATE keyword",
			input:    "UPDATE users SET",
			wantErr:  true,
			errMatch: "UPDATE",
		},
		{
			name:     "unbalanced single quotes",
			input:    "name = 'john",
			wantErr:  true,
			errMatch: "single quotes",
		},
		{
			name:     "unbalanced double quotes",
			input:    "name = \"john",
			wantErr:  true,
			errMatch: "double quotes",
		},
		{
			name:     "unbalanced parentheses open",
			input:    "id IN (1, 2, 3",
			wantErr:  true,
			errMatch: "parentheses",
		},
		{
			name:     "unbalanced parentheses close",
			input:    "id IN 1, 2, 3)",
			wantErr:  true,
			errMatch: "parentheses",
		},
		{
			name:     "no operator",
			input:    "user_id 123",
			wantErr:  true,
			errMatch: "operator",
		},
		{
			name:     "missing operator with keyword",
			input:    "status active",
			wantErr:  true,
			errMatch: "operator",
		},
		{
			name:    "complex valid query",
			input:   "user_id = 123 AND (status = 'active' OR status = 'pending')",
			wantErr: false,
		},
		{
			name:    "not equal operator",
			input:   "status != 'deleted'",
			wantErr: false,
		},
		{
			name:    "LIKE operator",
			input:   "name LIKE '%john%'",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilter(tt.input)
			if tt.wantErr {
				if err == "" {
					t.Errorf("validateFilter() expected error but got none")
				}
				if tt.errMatch != "" && !contains(err, tt.errMatch) {
					t.Errorf("validateFilter() error = %q, want error containing %q", err, tt.errMatch)
				}
			} else {
				if err != "" {
					t.Errorf("validateFilter() unexpected error = %q", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
