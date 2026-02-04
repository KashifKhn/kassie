package db

import (
	"context"
	"fmt"

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
	defer iter.Close()

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
	defer iter.Close()

	var results []map[string]interface{}
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		results = append(results, row)
	}

	nextPageState := iter.PageState()

	if err := iter.Close(); err != nil {
		return nil, nil, fmt.Errorf("query iteration failed: %w", err)
	}

	return results, nextPageState, nil
}

func (s *Session) Close() {
	if s.session != nil {
		s.session.Close()
	}
}

func (s *Session) Closed() bool {
	return s.session.Closed()
}
