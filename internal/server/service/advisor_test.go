package service

import (
	"context"
	"strings"
	"testing"

	pb "github.com/KashifKhn/kassie/api/gen/go"
)

func TestAnalyzeCompaction(t *testing.T) {
	stcs := map[string]interface{}{
		"class": "org.apache.cassandra.db.compaction.SizeTieredCompactionStrategy",
	}
	finding := analyzeCompaction("events", stcs)
	if finding == nil {
		t.Fatal("STCS should produce an info finding")
	}
	if finding.Severity != "info" || finding.Rule != "compaction-strategy" {
		t.Errorf("finding = %+v", finding)
	}
	if !strings.Contains(finding.Remediation, "LeveledCompactionStrategy") {
		t.Errorf("remediation missing strategy: %q", finding.Remediation)
	}

	lcs := map[string]interface{}{
		"class": "org.apache.cassandra.db.compaction.LeveledCompactionStrategy",
	}
	if analyzeCompaction("events", lcs) != nil {
		t.Error("LCS should not produce a finding")
	}

	if analyzeCompaction("events", nil) != nil {
		t.Error("nil compaction should not produce a finding")
	}
}

func TestAnalyzeTTL(t *testing.T) {
	finding := analyzeTTL("events", int32(0))
	if finding == nil {
		t.Fatal("ttl=0 should produce an info finding")
	}
	if finding.Rule != "no-default-ttl" {
		t.Errorf("rule = %q", finding.Rule)
	}

	if analyzeTTL("events", int32(86400)) != nil {
		t.Error("ttl set should not produce a finding")
	}
}

func TestAnalyzePartitionSize(t *testing.T) {
	stats := &pb.TableStats{
		EstimateAvailable:     true,
		MaxPartitionSizeBytes: 200 * 1024 * 1024,
	}
	finding := analyzePartitionSize("wide", stats)
	if finding == nil {
		t.Fatal("200MB partition should warn")
	}
	if finding.Severity != "warning" || finding.Rule != "large-partition" {
		t.Errorf("finding = %+v", finding)
	}
	if !strings.Contains(finding.Message, "200.0MB") {
		t.Errorf("size not humanized: %q", finding.Message)
	}

	small := &pb.TableStats{EstimateAvailable: true, MaxPartitionSizeBytes: 1024}
	if analyzePartitionSize("ok", small) != nil {
		t.Error("small partition should not warn")
	}

	noEstimate := &pb.TableStats{EstimateAvailable: false}
	if analyzePartitionSize("x", noEstimate) != nil {
		t.Error("no estimate should produce nothing")
	}
}

func TestAnalyzeTableSize(t *testing.T) {
	stats := &pb.TableStats{EstimateAvailable: true, RowCount: 250_000_000}
	finding := analyzeTableSize("huge", stats)
	if finding == nil || finding.Rule != "large-table" {
		t.Fatalf("finding = %+v", finding)
	}
	if !strings.Contains(finding.Message, "250.0M") {
		t.Errorf("count not humanized: %q", finding.Message)
	}

	if analyzeTableSize("small", &pb.TableStats{EstimateAvailable: true, RowCount: 1000}) != nil {
		t.Error("small table should not produce a finding")
	}
}

type fakeQuerier struct {
	rows []map[string]interface{}
}

func (f *fakeQuerier) FetchAll(ctx context.Context, stmt string, values ...interface{}) ([]map[string]interface{}, error) {
	return f.rows, nil
}

func TestAnalyzeReplication(t *testing.T) {
	simple := &fakeQuerier{rows: []map[string]interface{}{{
		"replication": map[string]interface{}{
			"class":              "org.apache.cassandra.locator.SimpleStrategy",
			"replication_factor": int64(1),
		},
	}}}

	findings := analyzeReplication(nil, simple, "ks")
	if len(findings) != 2 {
		t.Fatalf("findings = %d, want 2 (simple strategy + rf1)", len(findings))
	}
	rules := map[string]bool{}
	for _, f := range findings {
		rules[f.Rule] = true
		if f.Severity != "warning" {
			t.Errorf("severity = %q, want warning", f.Severity)
		}
	}
	if !rules["simple-strategy"] || !rules["replication-factor-1"] {
		t.Errorf("rules missing: %v", rules)
	}

	nts := &fakeQuerier{rows: []map[string]interface{}{{
		"replication": map[string]interface{}{
			"class":       "org.apache.cassandra.locator.NetworkTopologyStrategy",
			"datacenter1": int64(3),
		},
	}}}
	if got := analyzeReplication(nil, nts, "ks"); len(got) != 0 {
		t.Errorf("healthy NTS should produce no findings: %+v", got)
	}
}

func TestHumanBytesAndCount(t *testing.T) {
	if got := humanBytes(200 * 1024 * 1024); got != "200.0MB" {
		t.Errorf("humanBytes = %q", got)
	}
	if got := humanBytes(3 * 1024); got != "3.0KB" {
		t.Errorf("humanBytes = %q", got)
	}
	if got := humanCount(1_500_000); got != "1.5M" {
		t.Errorf("humanCount = %q", got)
	}
	if got := humanCount(2_300_000_000); got != "2.3B" {
		t.Errorf("humanCount = %q", got)
	}
}

func TestSortFindings(t *testing.T) {
	findings := []*pb.AdvisorFinding{
		{Severity: "info", Rule: "a"},
		{Severity: "warning", Rule: "b"},
		{Severity: "info", Rule: "c"},
	}
	sortFindings(findings)
	if findings[0].Severity != "warning" {
		t.Errorf("warnings should sort first: %+v", findings[0])
	}
}
