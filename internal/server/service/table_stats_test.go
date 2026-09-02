package service

import (
	"testing"

	pb "github.com/KashifKhn/kassie/api/gen/go"
)

func TestAggregateSizeEstimates(t *testing.T) {
	tests := []struct {
		name          string
		rows          []map[string]interface{}
		wantAvailable bool
		wantRows      int64
		wantMax       int64
	}{
		{
			name: "empty estimates unavailable",
			rows: nil,
		},
		{
			name: "all negative values unavailable",
			rows: []map[string]interface{}{
				{"partitions_count": int64(-1), "mean_partition_size": int64(-1)},
			},
		},
		{
			name: "nil values skipped",
			rows: []map[string]interface{}{
				{"partitions_count": nil, "mean_partition_size": nil},
				{"partitions_count": int64(10), "mean_partition_size": int64(200)},
			},
			wantAvailable: true,
			wantRows:      10,
			wantMax:       200,
		},
		{
			name: "sums across ranges",
			rows: []map[string]interface{}{
				{"partitions_count": int64(10), "mean_partition_size": int64(100)},
				{"partitions_count": int64(30), "mean_partition_size": int64(300)},
			},
			wantAvailable: true,
			wantRows:      40,
			wantMax:       300,
		},
		{
			name: "zero counts ignored but valid row counts",
			rows: []map[string]interface{}{
				{"partitions_count": int64(0), "mean_partition_size": int64(50)},
				{"partitions_count": int64(5), "mean_partition_size": int64(500)},
			},
			wantAvailable: true,
			wantRows:      5,
			wantMax:       500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &pb.TableStats{}
			aggregateSizeEstimates(tt.rows, stats)

			if stats.EstimateAvailable != tt.wantAvailable {
				t.Fatalf("estimate_available = %v, want %v", stats.EstimateAvailable, tt.wantAvailable)
			}
			if stats.RowCount != tt.wantRows {
				t.Errorf("row_count = %d, want %d", stats.RowCount, tt.wantRows)
			}
			if stats.MaxPartitionSizeBytes != tt.wantMax {
				t.Errorf("max_partition = %d, want %d", stats.MaxPartitionSizeBytes, tt.wantMax)
			}
		})
	}
}

func TestAsNonNegativeInt64(t *testing.T) {
	if got := asNonNegativeInt64(int64(-5)); got != 0 {
		t.Errorf("negative = %d, want 0", got)
	}
	if got := asNonNegativeInt64(int(7)); got != 7 {
		t.Errorf("int = %d, want 7", got)
	}
	if got := asNonNegativeInt64("nope"); got != 0 {
		t.Errorf("string = %d, want 0", got)
	}
	if got := asNonNegativeInt64(nil); got != 0 {
		t.Errorf("nil = %d, want 0", got)
	}
}
