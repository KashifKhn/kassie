package service

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/db"
	"github.com/gocql/gocql"
)

func convertTypedRows(page *db.Page) []*pb.Row {
	pbRows := make([]*pb.Row, 0, len(page.Rows))

	for _, row := range page.Rows {
		cells := make(map[string]*pb.CellValue, len(page.Columns))
		for i, col := range page.Columns {
			var cellType string
			if i < len(page.Types) {
				cellType = page.Types[i]
			}
			cells[col] = cellToPbValue(row[i], cellType)
		}
		pbRows = append(pbRows, &pb.Row{Cells: cells})
	}

	return pbRows
}

func cellToPbValue(value interface{}, cqlType string) *pb.CellValue {
	if value == nil {
		return &pb.CellValue{IsNull: true, CqlType: cqlType}
	}

	cell := &pb.CellValue{IsNull: false, CqlType: cqlType}

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
	case time.Time:
		cell.Value = &pb.CellValue_StringVal{StringVal: v.UTC().Format(time.RFC3339Nano)}
	case gocql.UUID:
		cell.Value = &pb.CellValue_StringVal{StringVal: v.String()}
	case net.IP:
		cell.Value = &pb.CellValue_StringVal{StringVal: v.String()}
	default:
		if data, err := json.Marshal(jsonReady(value)); err == nil {
			cell.Value = &pb.CellValue_StringVal{StringVal: string(data)}
		} else {
			cell.Value = &pb.CellValue_StringVal{StringVal: fmt.Sprintf("%v", value)}
		}
	}

	return cell
}

func jsonReady(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(v))
		for k, item := range v {
			out[k] = jsonReady(item)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(v))
		for i, item := range v {
			out[i] = jsonReady(item)
		}
		return out
	case gocql.UUID:
		return v.String()
	case net.IP:
		return v.String()
	case time.Time:
		return v.UTC().Format(time.RFC3339Nano)
	default:
		return value
	}
}
