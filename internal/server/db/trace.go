package db

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/gocql/gocql"
)

type TraceEvent struct {
	Activity  string
	Source    string
	ElapsedUs int64
	Thread    string
}

type TraceResult struct {
	Events      []TraceEvent
	DurationUs  int64
	Coordinator string
	Ready       bool
}

type idTracer struct {
	mu      sync.Mutex
	traceID []byte
}

func (t *idTracer) Trace(traceID []byte) {
	t.mu.Lock()
	t.traceID = traceID
	t.mu.Unlock()
}

func (t *idTracer) ID() []byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.traceID
}

func (s *Session) FetchTracedPage(ctx context.Context, stmt string, pageSize int, pageState []byte, values ...interface{}) (*Page, string, error) {
	tracer := &idTracer{}
	query := s.QueryContext(ctx, stmt, values...).PageSize(pageSize).Trace(tracer)
	if pageState != nil {
		query = query.PageState(pageState)
	}

	iter := query.Iter()

	page, err := s.collectPage(iter, pageSize)
	if err != nil {
		return nil, "", err
	}

	traceID := ""
	if id := tracer.ID(); id != nil {
		uuid, err := gocql.UUIDFromBytes(id)
		if err == nil {
			traceID = uuid.String()
		}
	}

	return page, traceID, nil
}

func (s *Session) collectPage(iter *gocql.Iter, pageSize int) (*Page, error) {
	cols := iter.Columns()
	columns := make([]string, len(cols))
	types := make([]string, len(cols))
	for i, c := range cols {
		columns[i] = c.Name
		types[i] = CQLTypeString(c.TypeInfo)
	}

	page := &Page{Columns: columns, Types: types}
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

func (s *Session) GetTrace(ctx context.Context, traceID string) (*TraceResult, error) {
	uuid, err := gocql.ParseUUID(traceID)
	if err != nil {
		return nil, fmt.Errorf("invalid trace id: %w", err)
	}

	dest := map[string]interface{}{}
	err = s.QueryContext(ctx,
		`SELECT coordinator, duration FROM system_traces.sessions WHERE session_id = ?`,
		uuid).Consistency(gocql.One).MapScan(dest)
	if err != nil {
		if err == gocql.ErrNotFound {
			return &TraceResult{Ready: false}, nil
		}
		return nil, fmt.Errorf("trace session lookup failed: %w", err)
	}

	coordinator := ""
	if v, ok := dest["coordinator"].(string); ok {
		coordinator = v
	}
	durationUs := int64(0)
	switch v := dest["duration"].(type) {
	case int64:
		durationUs = v
	case int:
		durationUs = int64(v)
	case int32:
		durationUs = int64(v)
	}

	iter := s.QueryContext(ctx,
		`SELECT activity, source, source_elapsed, thread FROM system_traces.events WHERE session_id = ?`,
		uuid).Consistency(gocql.One).Iter()

	var events []TraceEvent
	for {
		row := map[string]interface{}{}
		if !iter.MapScan(row) {
			break
		}
		event := TraceEvent{
			Activity: asString(row["activity"]),
			Source:   asString(row["source"]),
			Thread:   asString(row["thread"]),
		}
		if v, ok := row["source_elapsed"].(int64); ok {
			event.ElapsedUs = v
		}
		events = append(events, event)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("trace events iteration failed: %w", err)
	}

	return &TraceResult{
		Events:      events,
		DurationUs:  durationUs,
		Coordinator: coordinator,
		Ready:       true,
	}, nil
}

func asString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
