package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	largePartitionBytes = 100 * 1024 * 1024
	hugeTableRows       = 100_000_000
)

func (s *SchemaService) AnalyzeKeyspace(ctx context.Context, req *pb.AnalyzeKeyspaceRequest) (*pb.AnalyzeKeyspaceResponse, error) {
	if req.Keyspace == "" {
		return nil, status.Error(codes.InvalidArgument, "keyspace is required")
	}
	if err := validateIdentifier(req.Keyspace); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid keyspace: %v", err)
	}

	session, err := GetSessionFromContext(ctx, s.store)
	if err != nil {
		return nil, err
	}

	tableRows, err := session.Connection.FetchAll(ctx,
		`SELECT table_name, compaction, default_time_to_live FROM system_schema.tables WHERE keyspace_name = ?`,
		req.Keyspace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read table metadata: %v", err)
	}
	if len(tableRows) == 0 {
		return nil, status.Errorf(codes.NotFound, "keyspace %q has no tables", req.Keyspace)
	}

	var findings []*pb.AdvisorFinding

	for _, row := range tableRows {
		table := stringValue(row["table_name"])

		findings = append(findings,
			analyzeCompaction(table, row["compaction"]),
			analyzeTTL(table, row["default_time_to_live"]),
		)

		if stats, err := session.Connection.FetchAll(ctx,
			`SELECT mean_partition_size, partitions_count FROM system.size_estimates WHERE keyspace_name = ? AND table_name = ?`,
			req.Keyspace, table); err == nil && len(stats) > 0 {
			tableStats := &pb.TableStats{}
			aggregateSizeEstimates(stats, tableStats)
			findings = append(findings,
				analyzePartitionSize(table, tableStats),
				analyzeTableSize(table, tableStats),
			)
		}
	}

	replicationFindings := analyzeReplication(ctx, session.Connection, req.Keyspace)
	findings = append(findings, replicationFindings...)

	nonNil := make([]*pb.AdvisorFinding, 0, len(findings))
	for _, f := range findings {
		if f != nil {
			nonNil = append(nonNil, f)
		}
	}

	sortFindings(nonNil)

	return &pb.AnalyzeKeyspaceResponse{
		Findings:       nonNil,
		TablesAnalyzed: int32(len(tableRows)),
	}, nil
}

func analyzeCompaction(table string, compaction interface{}) *pb.AdvisorFinding {
	strategy := ""
	switch v := compaction.(type) {
	case map[string]interface{}:
		if cls, ok := v["class"].(string); ok {
			parts := strings.Split(cls, ".")
			strategy = parts[len(parts)-1]
		}
	case string:
		if strings.Contains(v, "CompactionStrategy") {
			strategy = v
		}
	}

	if strategy == "SizeTieredCompactionStrategy" {
		return &pb.AdvisorFinding{
			Severity: "info",
			Rule:     "compaction-strategy",
			Table:    table,
			Message:  fmt.Sprintf("uses %s (time-series workloads often prefer LeveledCompactionStrategy for read-heavy or TimeWindowCompactionStrategy for TTL'd data)", strategy),
			Remediation: fmt.Sprintf(
				`ALTER TABLE "%s" WITH compaction = {'class': 'LeveledCompactionStrategy'};`,
				table),
		}
	}
	return nil
}

func analyzeTTL(table string, ttl interface{}) *pb.AdvisorFinding {
	ttlSeconds := asNonNegativeInt64(ttl)
	if ttlSeconds == 0 {
		return &pb.AdvisorFinding{
			Severity: "info",
			Rule:     "no-default-ttl",
			Table:    table,
			Message:  "no default_time_to_live — rows persist forever; add TTL if data ages out",
			Remediation: fmt.Sprintf(
				`ALTER TABLE "%s" WITH default_time_to_live = 2592000;`,
				table),
		}
	}
	return nil
}

func analyzePartitionSize(table string, stats *pb.TableStats) *pb.AdvisorFinding {
	if !stats.EstimateAvailable {
		return nil
	}
	if stats.MaxPartitionSizeBytes >= largePartitionBytes {
		return &pb.AdvisorFinding{
			Severity: "warning",
			Rule:     "large-partition",
			Table:    table,
			Message: fmt.Sprintf("max partition ~%s approaches/exceeds the 100MB Cassandra guideline — risk of timeouts and GC pressure",
				humanBytes(stats.MaxPartitionSizeBytes)),
			Remediation: "review partition key cardinality or bucket large partitions (e.g. by time bucket)",
		}
	}
	return nil
}

func analyzeTableSize(table string, stats *pb.TableStats) *pb.AdvisorFinding {
	if !stats.EstimateAvailable {
		return nil
	}
	if stats.RowCount >= hugeTableRows {
		return &pb.AdvisorFinding{
			Severity:    "info",
			Rule:        "large-table",
			Table:       table,
			Message:     fmt.Sprintf("estimated %s rows — ensure partition-key queries and avoid full scans", humanCount(stats.RowCount)),
			Remediation: "",
		}
	}
	return nil
}

type sessionQuerier interface {
	FetchAll(ctx context.Context, stmt string, values ...interface{}) ([]map[string]interface{}, error)
}

func analyzeReplication(ctx context.Context, session sessionQuerier, keyspace string) []*pb.AdvisorFinding {
	rows, err := session.FetchAll(ctx,
		`SELECT replication FROM system_schema.keyspaces WHERE keyspace_name = ?`, keyspace)
	if err != nil || len(rows) == 0 {
		return nil
	}

	rep, ok := rows[0]["replication"].(map[string]interface{})
	if !ok {
		return nil
	}

	var findings []*pb.AdvisorFinding

	rf := 0
	for k, v := range rep {
		if strings.HasPrefix(strings.ToLower(k), "replication_factor") {
			if n := asNonNegativeInt64(v); n > 0 {
				rf = int(n)
			}
		}
	}

	strategy := stringValue(rep["class"])
	if strings.Contains(strategy, "SimpleStrategy") {
		findings = append(findings, &pb.AdvisorFinding{
			Severity: "warning",
			Rule:     "simple-strategy",
			Table:    "",
			Message:  "SimpleStrategy has no multi-DC awareness — use NetworkTopologyStrategy for production",
			Remediation: fmt.Sprintf(
				`ALTER KEYSPACE "%s" WITH replication = {'class': 'NetworkTopologyStrategy', 'datacenter1': %d};`,
				keyspace, maxInt(3, rf)),
		})
	}

	if rf == 1 {
		findings = append(findings, &pb.AdvisorFinding{
			Severity: "warning",
			Rule:     "replication-factor-1",
			Table:    "",
			Message:  "replication factor 1 — single node holds every row; data loss on node failure",
			Remediation: fmt.Sprintf(
				`ALTER KEYSPACE "%s" WITH replication = {'class': 'NetworkTopologyStrategy', 'datacenter1': 3};`,
				keyspace),
		})
	}

	return findings
}

func sortFindings(findings []*pb.AdvisorFinding) {
	rank := map[string]int{"warning": 0, "info": 1}
	sort.Slice(findings, func(i, j int) bool {
		return rank[findings[i].Severity] < rank[findings[j].Severity]
	})
}

func humanBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1fGB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1fMB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1fKB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%dB", n)
	}
}

func humanCount(n int64) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
