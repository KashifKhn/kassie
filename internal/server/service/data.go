package service

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/state"
)

type DataService struct {
	pb.UnimplementedDataServiceServer
	store *state.Store
}

func NewDataService(store *state.Store) *DataService {
	return &DataService{
		store: store,
	}
}

func (d *DataService) QueryRows(ctx context.Context, req *pb.QueryRowsRequest) (*pb.QueryRowsResponse, error) {
	if req.Keyspace == "" || req.Table == "" {
		return nil, fmt.Errorf("keyspace and table are required")
	}

	session, err := GetSessionFromContext(ctx, d.store)
	if err != nil {
		return nil, err
	}

	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 10000 {
		pageSize = 100
	}

	query := fmt.Sprintf("SELECT * FROM %s.%s", req.Keyspace, req.Table)
	rows, nextPageState, err := session.Connection.FetchWithPaging(ctx, query, pageSize, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}

	pbRows := make([]*pb.Row, 0, len(rows))
	for _, row := range rows {
		pbRow := rowToPbRow(row)
		pbRows = append(pbRows, pbRow)
	}

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
		return nil, fmt.Errorf("cursor ID is required")
	}

	session, err := GetSessionFromContext(ctx, d.store)
	if err != nil {
		return nil, err
	}

	cursor, err := session.Cursors.Get(req.CursorId)
	if err != nil {
		return nil, fmt.Errorf("cursor not found or expired: %w", err)
	}

	query := fmt.Sprintf("SELECT * FROM %s.%s", cursor.Keyspace, cursor.Table)
	if cursor.Filter != "" {
		query += " WHERE " + cursor.Filter
	}

	rows, nextPageState, err := session.Connection.FetchWithPaging(ctx, query, cursor.PageSize, cursor.PageState)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch next page: %w", err)
	}

	pbRows := make([]*pb.Row, 0, len(rows))
	for _, row := range rows {
		pbRow := rowToPbRow(row)
		pbRows = append(pbRows, pbRow)
	}

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
		return nil, fmt.Errorf("keyspace and table are required")
	}

	if req.WhereClause == "" {
		return nil, fmt.Errorf("where clause is required for filtering")
	}

	session, err := GetSessionFromContext(ctx, d.store)
	if err != nil {
		return nil, err
	}

	if err := validateWhereClause(req.WhereClause); err != nil {
		return nil, fmt.Errorf("invalid WHERE clause: %w", err)
	}

	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 10000 {
		pageSize = 100
	}

	query := fmt.Sprintf("SELECT * FROM %s.%s WHERE %s", req.Keyspace, req.Table, req.WhereClause)
	rows, nextPageState, err := session.Connection.FetchWithPaging(ctx, query, pageSize, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to filter rows: %w", err)
	}

	pbRows := make([]*pb.Row, 0, len(rows))
	for _, row := range rows {
		pbRow := rowToPbRow(row)
		pbRows = append(pbRows, pbRow)
	}

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

func validateWhereClause(whereClause string) error {
	whereClause = strings.TrimSpace(strings.ToLower(whereClause))

	dangerousKeywords := []string{"drop", "delete", "insert", "update", "alter", "create", "truncate"}
	for _, keyword := range dangerousKeywords {
		if strings.Contains(whereClause, keyword) {
			return fmt.Errorf("dangerous keyword detected: %s", keyword)
		}
	}

	if whereClause == "" {
		return fmt.Errorf("empty WHERE clause")
	}

	return nil
}
