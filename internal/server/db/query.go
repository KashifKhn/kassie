package db

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	keyspace string
	table    string
}

func NewQueryBuilder(keyspace, table string) *QueryBuilder {
	return &QueryBuilder{
		keyspace: keyspace,
		table:    table,
	}
}

func (qb *QueryBuilder) SelectAll(limit int) string {
	query := fmt.Sprintf("SELECT * FROM %s.%s", qb.keyspace, qb.table)
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	return query
}

func (qb *QueryBuilder) SelectWithWhere(whereClause string, limit int) string {
	query := fmt.Sprintf("SELECT * FROM %s.%s WHERE %s", qb.keyspace, qb.table, whereClause)
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	return query
}

func (qb *QueryBuilder) Count() string {
	return fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", qb.keyspace, qb.table)
}

func (qb *QueryBuilder) Insert(columns []string) string {
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	return fmt.Sprintf(
		"INSERT INTO %s.%s (%s) VALUES (%s)",
		qb.keyspace,
		qb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
}

func (qb *QueryBuilder) Update(columns []string, whereClause string) string {
	setClauses := make([]string, len(columns))
	for i, col := range columns {
		setClauses[i] = fmt.Sprintf("%s = ?", col)
	}

	return fmt.Sprintf(
		"UPDATE %s.%s SET %s WHERE %s",
		qb.keyspace,
		qb.table,
		strings.Join(setClauses, ", "),
		whereClause,
	)
}

func (qb *QueryBuilder) Delete(whereClause string) string {
	return fmt.Sprintf("DELETE FROM %s.%s WHERE %s", qb.keyspace, qb.table, whereClause)
}

func ListKeyspaces() string {
	return "SELECT keyspace_name FROM system_schema.keyspaces"
}

func ListTables(keyspace string) string {
	return fmt.Sprintf("SELECT table_name FROM system_schema.tables WHERE keyspace_name = '%s'", keyspace)
}

func DescribeTable(keyspace, table string) string {
	return fmt.Sprintf(
		"SELECT column_name, type, kind FROM system_schema.columns WHERE keyspace_name = '%s' AND table_name = '%s'",
		keyspace,
		table,
	)
}
