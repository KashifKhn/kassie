package db

import (
	"fmt"
	"regexp"
	"strings"
)

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func QuoteIdentifier(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("identifier cannot be empty")
	}
	if !validIdentifier.MatchString(name) {
		return "", fmt.Errorf("invalid identifier: %q", name)
	}
	return `"` + name + `"`, nil
}

type QueryBuilder struct {
	keyspace string
	table    string
}

func NewQueryBuilder(keyspace, table string) (*QueryBuilder, error) {
	if keyspace == "" || table == "" {
		return nil, fmt.Errorf("keyspace and table must not be empty")
	}
	if !validIdentifier.MatchString(keyspace) {
		return nil, fmt.Errorf("invalid keyspace identifier: %q", keyspace)
	}
	if !validIdentifier.MatchString(table) {
		return nil, fmt.Errorf("invalid table identifier: %q", table)
	}
	return &QueryBuilder{keyspace: keyspace, table: table}, nil
}

func (qb *QueryBuilder) qualifiedTable() string {
	return fmt.Sprintf(`"%s"."%s"`, qb.keyspace, qb.table)
}

func (qb *QueryBuilder) SelectAll(limit int) string {
	query := fmt.Sprintf("SELECT * FROM %s", qb.qualifiedTable())
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	return query
}

func (qb *QueryBuilder) SelectWithWhere(whereClause string, limit int) string {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", qb.qualifiedTable(), whereClause)
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	return query
}

func (qb *QueryBuilder) Count() string {
	return fmt.Sprintf("SELECT COUNT(*) FROM %s", qb.qualifiedTable())
}

func (qb *QueryBuilder) Insert(columns []string) (string, error) {
	if len(columns) == 0 {
		return "", fmt.Errorf("at least one column is required for insert")
	}

	quoted := make([]string, len(columns))
	for i, col := range columns {
		q, err := QuoteIdentifier(col)
		if err != nil {
			return "", fmt.Errorf("invalid column name: %w", err)
		}
		quoted[i] = q
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		qb.qualifiedTable(),
		strings.Join(quoted, ", "),
		strings.Join(placeholders, ", "),
	), nil
}

func (qb *QueryBuilder) Update(columns []string, whereClause string) (string, error) {
	if len(columns) == 0 {
		return "", fmt.Errorf("at least one column is required for update")
	}

	setClauses := make([]string, len(columns))
	for i, col := range columns {
		q, err := QuoteIdentifier(col)
		if err != nil {
			return "", fmt.Errorf("invalid column name: %w", err)
		}
		setClauses[i] = fmt.Sprintf("%s = ?", q)
	}

	return fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		qb.qualifiedTable(),
		strings.Join(setClauses, ", "),
		whereClause,
	), nil
}

func (qb *QueryBuilder) Delete(whereClause string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s", qb.qualifiedTable(), whereClause)
}

func ListKeyspaces() string {
	return "SELECT keyspace_name FROM system_schema.keyspaces"
}

func ListTables(keyspace string) (string, error) {
	if !validIdentifier.MatchString(keyspace) {
		return "", fmt.Errorf("invalid keyspace identifier: %q", keyspace)
	}
	return "SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?", nil
}

func DescribeTable(keyspace, table string) (string, error) {
	if !validIdentifier.MatchString(keyspace) {
		return "", fmt.Errorf("invalid keyspace identifier: %q", keyspace)
	}
	if !validIdentifier.MatchString(table) {
		return "", fmt.Errorf("invalid table identifier: %q", table)
	}
	return "SELECT column_name, type, kind FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ?", nil
}
