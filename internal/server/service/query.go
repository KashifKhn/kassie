package service

import (
	"context"

	"fmt"
	"github.com/KashifKhn/kassie/internal/server/db"
	"regexp"
	"strings"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxCQLLength = 10 * 1024
const maxQueryPageSize = 1000

var forbiddenQueryKeywords = regexp.MustCompile(
	`(?i)\b(INSERT|UPDATE|DELETE|ALTER|DROP|CREATE|TRUNCATE|GRANT|REVOKE|BATCH|USE|BEGIN|APPLY)\b`,
)

var selectOnlyPrefix = regexp.MustCompile(`(?i)^\s*SELECT\s`)

func (d *DataService) ExecuteQuery(ctx context.Context, req *pb.ExecuteQueryRequest) (*pb.ExecuteQueryResponse, error) {
	cql, err := normalizeAdhocCQL(req.Cql)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	session, err := GetSessionFromContext(ctx, d.store)
	if err != nil {
		return nil, err
	}

	pageSize := normalizePageSize(int(req.PageSize))
	if pageSize > maxQueryPageSize {
		pageSize = maxQueryPageSize
	}

	start := time.Now()
	var page *db.Page
	var traceID string
	if req.Trace {
		page, traceID, err = session.Connection.FetchTracedPage(ctx, cql, pageSize, nil)
	} else {
		page, err = session.Connection.FetchPage(ctx, cql, pageSize, nil)
	}
	elapsed := time.Since(start)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execute query: %v", err)
	}

	if d.queries != nil {
		d.queries.RecordLatency(session.Profile.Name, cql, elapsed.Milliseconds())
	}

	pbRows := convertTypedRows(page)
	nextPageState := page.NextPageState

	var cursorID string
	hasMore := len(nextPageState) > 0

	if hasMore {
		cursorID = session.Cursors.CreateWithCQL(nextPageState, cql, pageSize)
	}

	return &pb.ExecuteQueryResponse{
		Rows:         pbRows,
		CursorId:     cursorID,
		HasMore:      hasMore,
		TotalFetched: int64(len(page.Rows)),
		TraceId:      traceID,
	}, nil
}

func (d *DataService) GetTrace(ctx context.Context, req *pb.GetTraceRequest) (*pb.GetTraceResponse, error) {
	if req.TraceId == "" {
		return nil, status.Error(codes.InvalidArgument, "trace id is required")
	}

	session, err := GetSessionFromContext(ctx, d.store)
	if err != nil {
		return nil, err
	}

	trace, err := session.Connection.GetTrace(ctx, req.TraceId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to load trace: %v", err)
	}

	resp := &pb.GetTraceResponse{
		Ready:       trace.Ready,
		DurationUs:  trace.DurationUs,
		Coordinator: trace.Coordinator,
	}
	for _, e := range trace.Events {
		resp.Events = append(resp.Events, &pb.TraceEvent{
			Activity:  e.Activity,
			Source:    e.Source,
			ElapsedUs: e.ElapsedUs,
			Thread:    e.Thread,
		})
	}

	return resp, nil
}

func normalizeAdhocCQL(cql string) (string, error) {
	trimmed := strings.TrimSpace(cql)
	if trimmed == "" {
		return "", fmt.Errorf("query is required")
	}
	if len(trimmed) > maxCQLLength {
		return "", fmt.Errorf("query exceeds maximum length of %d characters", maxCQLLength)
	}
	if strings.Contains(trimmed, ";") {
		body := strings.TrimRight(trimmed, "; \t\n\r")
		if strings.Contains(body, ";") {
			return "", fmt.Errorf("multiple statements are not allowed")
		}
		trimmed = body
	}

	if !selectOnlyPrefix.MatchString(trimmed) {
		return "", fmt.Errorf("only SELECT statements are allowed")
	}
	if forbiddenQueryKeywords.MatchString(trimmed) {
		return "", fmt.Errorf("query contains disallowed keywords; only read-only SELECT statements are permitted")
	}

	return trimmed, nil
}
