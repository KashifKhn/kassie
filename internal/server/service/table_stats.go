package service

import (
	"context"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SchemaService) GetTableStats(ctx context.Context, req *pb.GetTableStatsRequest) (*pb.GetTableStatsResponse, error) {
	if req.Keyspace == "" || req.Table == "" {
		return nil, status.Error(codes.InvalidArgument, "keyspace and table are required")
	}

	if err := validateIdentifier(req.Keyspace); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid keyspace: %v", err)
	}
	if err := validateIdentifier(req.Table); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid table: %v", err)
	}

	session, err := GetSessionFromContext(ctx, s.store)
	if err != nil {
		return nil, err
	}

	stats := &pb.TableStats{}

	rows, err := session.Connection.FetchAll(ctx,
		`SELECT mean_partition_size, partitions_count FROM system.size_estimates WHERE keyspace_name = ? AND table_name = ?`,
		req.Keyspace, req.Table)
	if err == nil && len(rows) > 0 {
		aggregateSizeEstimates(rows, stats)
	}

	if !stats.EstimateAvailable {
		dest := map[string]interface{}{}
		if err := session.Connection.FetchOne(ctx, dest,
			`SELECT COUNT(*) AS row_count FROM "`+req.Keyspace+`"."`+req.Table+`"`); err == nil {
			if v, ok := dest["row_count"].(int64); ok {
				stats.RowCount = v
			}
		}
	}

	return &pb.GetTableStatsResponse{Stats: stats}, nil
}

func aggregateSizeEstimates(rows []map[string]interface{}, stats *pb.TableStats) {
	var totalPartitions int64
	var partitionCount int
	var maxPartitionSize int64

	for _, row := range rows {
		count := asNonNegativeInt64(row["partitions_count"])
		size := asNonNegativeInt64(row["mean_partition_size"])

		if count > 0 {
			totalPartitions += count
			partitionCount++
		}
		if size > maxPartitionSize {
			maxPartitionSize = size
		}
	}

	if partitionCount == 0 {
		return
	}

	stats.EstimateAvailable = true
	stats.RowCount = totalPartitions
	stats.MaxPartitionSizeBytes = maxPartitionSize
	if totalPartitions > 0 {
		stats.MeanPartitionSizeBytes = maxPartitionSize / int64(partitionCount)
	} else {
		stats.MeanPartitionSizeBytes = 0
	}
}

func asNonNegativeInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		if n > 0 {
			return n
		}
	case int:
		if n > 0 {
			return int64(n)
		}
	}
	return 0
}
