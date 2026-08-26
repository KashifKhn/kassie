package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/gocql/gocql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const exportChunkBytes = 512 * 1024

type rowFormatter func(columns []string, row []interface{}) ([]byte, error)

func (d *DataService) ExportRows(req *pb.ExportRowsRequest, stream pb.DataService_ExportRowsServer) error {
	if req.Keyspace == "" || req.Table == "" {
		return status.Error(codes.InvalidArgument, "keyspace and table are required")
	}

	if err := validateIdentifier(req.Keyspace); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid keyspace: %v", err)
	}
	if err := validateIdentifier(req.Table); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid table: %v", err)
	}

	if req.WhereClause != "" {
		if err := validateWhereClause(req.WhereClause); err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid WHERE clause: %v", err)
		}
	}

	session, err := GetSessionFromContext(stream.Context(), d.store)
	if err != nil {
		return err
	}

	format := req.Format
	if format == pb.ExportFormat_EXPORT_FORMAT_UNSPECIFIED {
		format = pb.ExportFormat_EXPORT_FORMAT_CSV
	}

	formatRow, writeHeader, err := formatterFor(format)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "%v", err)
	}

	fetchSize := normalizePageSize(int(req.FetchSize))

	query := fmt.Sprintf(`SELECT * FROM "%s"."%s"`, req.Keyspace, req.Table)
	if req.WhereClause != "" {
		query += " WHERE " + req.WhereClause
	}

	buf := &bytes.Buffer{}
	var rowsExported int64
	headerPending := true
	pageState := []byte{}

	send := func(done bool) error {
		if buf.Len() == 0 && !done {
			return nil
		}
		chunk := &pb.ExportChunk{
			Data:         buf.Bytes(),
			RowsExported: rowsExported,
			Done:         done,
		}
		buf.Reset()
		return stream.Send(chunk)
	}

	for {
		page, err := session.Connection.FetchPage(stream.Context(), query, fetchSize, pageState)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to fetch export page: %v", err)
		}

		if len(page.Columns) == 0 {
			break
		}

		if headerPending {
			header, err := writeHeader(page.Columns)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to write header: %v", err)
			}
			buf.Write(header)
			headerPending = false
		}

		for _, row := range page.Rows {
			line, err := formatRow(page.Columns, row)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to format row: %v", err)
			}
			buf.Write(line)
			rowsExported++
		}

		if buf.Len() >= exportChunkBytes {
			if err := send(false); err != nil {
				return err
			}
		}

		pageState = page.NextPageState
		if len(pageState) == 0 {
			break
		}
	}

	if err := send(true); err != nil {
		return err
	}

	return nil
}

func formatterFor(format pb.ExportFormat) (rowFormatter, func([]string) ([]byte, error), error) {
	switch format {
	case pb.ExportFormat_EXPORT_FORMAT_CSV:
		return csvRow, csvHeader, nil
	case pb.ExportFormat_EXPORT_FORMAT_JSON:
		return jsonRow, jsonHeader, nil
	default:
		return nil, nil, fmt.Errorf("unsupported export format")
	}
}

func csvHeader(columns []string) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err := w.Write(columns); err != nil {
		return nil, err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func csvRow(columns []string, row []interface{}) ([]byte, error) {
	record := make([]string, len(row))
	for i, v := range row {
		record[i] = formatCellCSV(v)
	}

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err := w.Write(record); err != nil {
		return nil, err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func jsonHeader([]string) ([]byte, error) { return nil, nil }

func jsonRow(columns []string, row []interface{}) ([]byte, error) {
	obj := make(map[string]interface{}, len(columns))
	for i, col := range columns {
		obj[col] = nativeValue(row[i])
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func formatCellCSV(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case time.Time:
		return val.UTC().Format(time.RFC3339Nano)
	case gocql.UUID:
		return val.String()
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func nativeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case gocql.UUID:
		return val.String()
	default:
		return v
	}
}
