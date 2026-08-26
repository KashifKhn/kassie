package db

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gocql/gocql"
)

type Session struct {
	session *gocql.Session
}

func NewSession(session *gocql.Session) *Session {
	return &Session{session: session}
}

func (s *Session) QueryContext(ctx context.Context, stmt string, values ...interface{}) *gocql.Query {
	return s.session.Query(stmt, values...).WithContext(ctx)
}

func (s *Session) ExecuteQuery(ctx context.Context, stmt string, values ...interface{}) error {
	return s.QueryContext(ctx, stmt, values...).Exec()
}

func (s *Session) FetchOne(ctx context.Context, dest map[string]interface{}, stmt string, values ...interface{}) error {
	return s.QueryContext(ctx, stmt, values...).MapScan(dest)
}

func (s *Session) FetchAll(ctx context.Context, stmt string, values ...interface{}) ([]map[string]interface{}, error) {
	iter := s.QueryContext(ctx, stmt, values...).Iter()

	var results []map[string]interface{}
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		results = append(results, row)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("query iteration failed: %w", err)
	}

	return results, nil
}

func (s *Session) FetchWithPaging(ctx context.Context, stmt string, pageSize int, pageState []byte, values ...interface{}) ([]map[string]interface{}, []byte, error) {
	query := s.QueryContext(ctx, stmt, values...).PageSize(pageSize)
	if pageState != nil {
		query = query.PageState(pageState)
	}

	iter := query.Iter()

	results := make([]map[string]interface{}, 0, pageSize)
	for len(results) < pageSize {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		results = append(results, row)
	}

	nextPageState := iter.PageState()
	if nextPageState != nil {
		pageStateCopy := make([]byte, len(nextPageState))
		copy(pageStateCopy, nextPageState)
		nextPageState = pageStateCopy
	}

	if err := iter.Close(); err != nil {
		return nil, nil, fmt.Errorf("query iteration failed: %w", err)
	}

	return results, nextPageState, nil
}

type Page struct {
	Columns       []string
	Rows          [][]interface{}
	NextPageState []byte
}

func (s *Session) FetchPage(ctx context.Context, stmt string, pageSize int, pageState []byte, values ...interface{}) (*Page, error) {
	query := s.QueryContext(ctx, stmt, values...).PageSize(pageSize)
	if pageState != nil {
		query = query.PageState(pageState)
	}

	iter := query.Iter()

	cols := iter.Columns()
	columns := make([]string, len(cols))
	for i, c := range cols {
		columns[i] = c.Name
	}

	page := &Page{Columns: columns}
	for len(page.Rows) < pageSize {
		rd, err := iter.RowData()
		if err != nil {
			return nil, fmt.Errorf("row data unavailable: %w", err)
		}
		if !iter.Scan(rd.Values...) {
			break
		}
		row := make([]interface{}, len(rd.Values))
		for i, v := range rd.Values {
			row[i] = reflect.Indirect(reflect.ValueOf(v)).Interface()
		}
		page.Rows = append(page.Rows, row)
	}

	nextPageState := iter.PageState()
	if nextPageState != nil {
		stateCopy := make([]byte, len(nextPageState))
		copy(stateCopy, nextPageState)
		page.NextPageState = stateCopy
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("query iteration failed: %w", err)
	}

	return page, nil
}

func (s *Session) Close() {
	if s.session != nil {
		s.session.Close()
	}
}

func (s *Session) Closed() bool {
	if s.session == nil {
		return true
	}
	return s.session.Closed()
}
