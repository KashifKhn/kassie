package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DataService struct {
	pb.UnimplementedDataServiceServer
	store SessionStore
}

func NewDataService(store SessionStore) *DataService {
	return &DataService{
		store: store,
	}
}

func (d *DataService) QueryRows(ctx context.Context, req *pb.QueryRowsRequest) (*pb.QueryRowsResponse, error) {
	if req.Keyspace == "" || req.Table == "" {
		return nil, status.Error(codes.InvalidArgument, "keyspace and table are required")
	}

	if err := validateIdentifier(req.Keyspace); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid keyspace: %v", err)
	}
	if err := validateIdentifier(req.Table); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid table: %v", err)
	}

	session, err := GetSessionFromContext(ctx, d.store)
	if err != nil {
		return nil, err
	}

	pageSize := normalizePageSize(int(req.PageSize))

	query := fmt.Sprintf(`SELECT * FROM "%s"."%s"`, req.Keyspace, req.Table)
	rows, nextPageState, err := session.Connection.FetchWithPaging(ctx, query, pageSize, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query rows: %v", err)
	}

	pbRows := convertRows(rows)

	var cursorID string
	hasMore := len(nextPageState) > 0

	if hasMore {
		cursorID = session.Cursors.Create(nextPageState, req.Keyspace, req.Table, "", pageSize)
	}

	return &pb.QueryRowsResponse{
		Rows:         pbRows,
		CursorId:     cursorID,
		HasMore:      hasMore,
		TotalFetched: int64(len(rows)),
	}, nil
}

func (d *DataService) GetNextPage(ctx context.Context, req *pb.GetNextPageRequest) (*pb.GetNextPageResponse, error) {
	if req.CursorId == "" {
		return nil, status.Error(codes.InvalidArgument, "cursor ID is required")
	}

	session, err := GetSessionFromContext(ctx, d.store)
	if err != nil {
		return nil, err
	}

	cursor, err := session.Cursors.Get(req.CursorId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cursor not found or expired: %v", err)
	}

	query := fmt.Sprintf(`SELECT * FROM "%s"."%s"`, cursor.Keyspace, cursor.Table)
	if cursor.Filter != "" {
		query += " WHERE " + cursor.Filter
	}

	rows, nextPageState, err := session.Connection.FetchWithPaging(ctx, query, cursor.PageSize, cursor.PageState)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch next page: %v", err)
	}

	pbRows := convertRows(rows)

	var newCursorID string
	hasMore := len(nextPageState) > 0

	if hasMore {
		newCursorID = session.Cursors.Create(nextPageState, cursor.Keyspace, cursor.Table, cursor.Filter, cursor.PageSize)
	}

	session.Cursors.Delete(req.CursorId)

	return &pb.GetNextPageResponse{
		Rows:     pbRows,
		CursorId: newCursorID,
		HasMore:  hasMore,
	}, nil
}

func (d *DataService) FilterRows(ctx context.Context, req *pb.FilterRowsRequest) (*pb.FilterRowsResponse, error) {
	if req.Keyspace == "" || req.Table == "" {
		return nil, status.Error(codes.InvalidArgument, "keyspace and table are required")
	}

	if req.WhereClause == "" {
		return nil, status.Error(codes.InvalidArgument, "where clause is required for filtering")
	}

	if err := validateIdentifier(req.Keyspace); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid keyspace: %v", err)
	}
	if err := validateIdentifier(req.Table); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid table: %v", err)
	}

	session, err := GetSessionFromContext(ctx, d.store)
	if err != nil {
		return nil, err
	}

	if err := validateWhereClause(req.WhereClause); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid WHERE clause: %v", err)
	}

	pageSize := normalizePageSize(int(req.PageSize))

	query := fmt.Sprintf(`SELECT * FROM "%s"."%s" WHERE %s`, req.Keyspace, req.Table, req.WhereClause)
	rows, nextPageState, err := session.Connection.FetchWithPaging(ctx, query, pageSize, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to filter rows: %v", err)
	}

	pbRows := convertRows(rows)

	var cursorID string
	hasMore := len(nextPageState) > 0

	if hasMore {
		cursorID = session.Cursors.Create(nextPageState, req.Keyspace, req.Table, req.WhereClause, pageSize)
	}

	return &pb.FilterRowsResponse{
		Rows:     pbRows,
		CursorId: cursorID,
		HasMore:  hasMore,
	}, nil
}

func convertRows(rows []map[string]interface{}) []*pb.Row {
	pbRows := make([]*pb.Row, 0, len(rows))
	for _, row := range rows {
		pbRows = append(pbRows, rowToPbRow(row))
	}
	return pbRows
}

func rowToPbRow(row map[string]interface{}) *pb.Row {
	cells := make(map[string]*pb.CellValue)

	for key, value := range row {
		cells[key] = interfaceToCellValue(value)
	}

	return &pb.Row{
		Cells: cells,
	}
}

func interfaceToCellValue(value interface{}) *pb.CellValue {
	if value == nil {
		return &pb.CellValue{IsNull: true}
	}

	cell := &pb.CellValue{IsNull: false}

	switch v := value.(type) {
	case string:
		cell.Value = &pb.CellValue_StringVal{StringVal: v}
	case int:
		cell.Value = &pb.CellValue_IntVal{IntVal: int64(v)}
	case int32:
		cell.Value = &pb.CellValue_IntVal{IntVal: int64(v)}
	case int64:
		cell.Value = &pb.CellValue_IntVal{IntVal: v}
	case float32:
		cell.Value = &pb.CellValue_DoubleVal{DoubleVal: float64(v)}
	case float64:
		cell.Value = &pb.CellValue_DoubleVal{DoubleVal: v}
	case bool:
		cell.Value = &pb.CellValue_BoolVal{BoolVal: v}
	case []byte:
		cell.Value = &pb.CellValue_BytesVal{BytesVal: v}
	default:
		cell.Value = &pb.CellValue_StringVal{StringVal: fmt.Sprintf("%v", v)}
	}

	return cell
}

var (
	identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

	dangerousStatements = regexp.MustCompile(
		`(?i)\b(DROP|DELETE\s+FROM|INSERT\s+INTO|UPDATE\s+\w+\s+SET|ALTER|CREATE|TRUNCATE|GRANT|REVOKE|BATCH)\b`,
	)

	commentPattern = regexp.MustCompile(`/\*|\*/|--`)
	controlChars   = regexp.MustCompile(`[\x00\n\r]`)
	cqlOperator    = regexp.MustCompile(`(?i)(!=|<=|>=|=|<|>|\bIN\b|\bCONTAINS\b)`)
)

func validateIdentifier(name string) error {
	if !identifierRegex.MatchString(name) {
		return status.Errorf(codes.InvalidArgument, "contains invalid characters: %q", name)
	}
	return nil
}

func normalizePageSize(pageSize int) int {
	if pageSize <= 0 || pageSize > 10000 {
		return 100
	}
	return pageSize
}

func validateWhereClause(whereClause string) error {
	trimmed := strings.TrimSpace(whereClause)
	if trimmed == "" {
		return status.Error(codes.InvalidArgument, "empty WHERE clause")
	}

	if strings.Contains(trimmed, ";") {
		return status.Error(codes.InvalidArgument, "semicolons are not allowed in WHERE clause")
	}

	if commentPattern.MatchString(trimmed) {
		return status.Error(codes.InvalidArgument, "comments are not allowed in WHERE clause")
	}

	if controlChars.MatchString(trimmed) {
		return status.Error(codes.InvalidArgument, "control characters are not allowed in WHERE clause")
	}

	if dangerousStatements.MatchString(trimmed) {
		return status.Error(codes.InvalidArgument, "WHERE clause contains disallowed CQL statement")
	}

	if !cqlOperator.MatchString(trimmed) {
		return status.Error(codes.InvalidArgument, "WHERE clause must contain at least one comparison operator")
	}

	return nil
}

func ValidateIdentifier(name string) error {
	_, err := db.QuoteIdentifier(name)
	return err
}
