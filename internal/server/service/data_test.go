package service

import "testing"

func TestValidateWhereClause(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid equality", input: "id = 1", wantErr: false},
		{name: "valid string comparison", input: "name = 'alice'", wantErr: false},
		{name: "valid less than", input: "age < 30", wantErr: false},
		{name: "valid greater than or equal", input: "age >= 18", wantErr: false},
		{name: "valid IN clause", input: "id IN (1, 2, 3)", wantErr: false},
		{name: "valid CONTAINS", input: "tags CONTAINS 'go'", wantErr: false},
		{name: "valid compound AND", input: "id = 1 AND name = 'bob'", wantErr: false},
		{name: "valid not equal", input: "status != 'active'", wantErr: false},
		{name: "empty clause", input: "", wantErr: true},
		{name: "whitespace only", input: "   ", wantErr: true},
		{name: "semicolon injection", input: "id = 1; DROP TABLE users", wantErr: true},
		{name: "block comment open", input: "id = 1 /* comment", wantErr: true},
		{name: "block comment close", input: "comment */ id = 1", wantErr: true},
		{name: "line comment", input: "id = 1 -- comment", wantErr: true},
		{name: "newline injection", input: "id = 1\nDROP TABLE users", wantErr: true},
		{name: "carriage return", input: "id = 1\rDROP TABLE users", wantErr: true},
		{name: "null byte", input: "id = 1\x00", wantErr: true},
		{name: "DROP statement", input: "DROP TABLE users", wantErr: true},
		{name: "DELETE FROM", input: "DELETE FROM users WHERE id=1", wantErr: true},
		{name: "INSERT INTO", input: "INSERT INTO users VALUES (1)", wantErr: true},
		{name: "ALTER statement", input: "ALTER TABLE users ADD col text", wantErr: true},
		{name: "TRUNCATE statement", input: "TRUNCATE users", wantErr: true},
		{name: "no operator", input: "just some text", wantErr: true},
		{name: "identifier only", input: "id", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWhereClause(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWhereClause(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "simple name", input: "users", wantErr: false},
		{name: "with underscore", input: "my_table", wantErr: false},
		{name: "starts with underscore", input: "_private", wantErr: false},
		{name: "with numbers", input: "table1", wantErr: false},
		{name: "uppercase", input: "MyTable", wantErr: false},
		{name: "contains space", input: "my table", wantErr: true},
		{name: "contains dot", input: "my.table", wantErr: true},
		{name: "contains semicolon", input: "table;drop", wantErr: true},
		{name: "starts with number", input: "1table", wantErr: true},
		{name: "contains quotes", input: `my"table`, wantErr: true},
		{name: "empty string", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIdentifier(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIdentifier(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizePageSize(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{name: "zero", in: 0, want: 100},
		{name: "negative", in: -5, want: 100},
		{name: "over max", in: 10001, want: 100},
		{name: "valid small", in: 1, want: 1},
		{name: "valid normal", in: 50, want: 50},
		{name: "valid max", in: 10000, want: 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePageSize(tt.in)
			if got != tt.want {
				t.Errorf("normalizePageSize(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
