//go:build integration

package grpc_test

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/client"
	serverdb "github.com/KashifKhn/kassie/internal/server/db"
	servergrpc "github.com/KashifKhn/kassie/internal/server/grpc"
	"github.com/KashifKhn/kassie/internal/server/state"
	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/KashifKhn/kassie/internal/shared/logger"
	"github.com/gocql/gocql"
)

const exportRowCount = 250
const exportKeyspace = "kassie_export_it"

type stubProfileProvider struct {
	profile *config.Profile
}

func (p *stubProfileProvider) GetProfile(name string) (*config.Profile, error) {
	if name == p.profile.Name {
		return p.profile, nil
	}
	return nil, fmt.Errorf("profile not found: %s", name)
}

func (p *stubProfileProvider) GetProfiles() []config.Profile {
	return []config.Profile{*p.profile}
}

func testLogger(t *testing.T) *logger.Logger {
	t.Helper()
	log, err := logger.New(logger.Config{Level: logger.ErrorLevel, Pretty: false})
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	return log
}

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func cassandraAddr(t *testing.T) (string, int) {
	t.Helper()
	port := 9042
	if v := os.Getenv("KASSIE_TEST_CASSANDRA_PORT"); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &port); err != nil {
			t.Fatalf("invalid KASSIE_TEST_CASSANDRA_PORT %q", v)
		}
	}
	return envDefault("KASSIE_TEST_CASSANDRA_HOST", "127.0.0.1"), port
}

func startRealServer(t *testing.T) string {
	t.Helper()

	host, port := cassandraAddr(t)

	store := state.NewStore(time.Hour)
	pool := serverdb.NewPool()
	provider := &stubProfileProvider{profile: &config.Profile{
		Name:  "it-export",
		Hosts: []string{host},
		Port:  port,
	}}

	cfg := &servergrpc.ServerConfig{
		Host:      "127.0.0.1",
		Port:      0,
		JWTSecret: "integration-test-secret",
	}

	srv, err := servergrpc.NewServer(cfg, &servergrpc.ServerDeps{
		Config:  provider,
		Pool:    pool,
		Store:   store,
		Queries: state.NewQueryStore(filepath.Join(t.TempDir(), "queries.json")),
	}, testLogger(t))
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	if err := srv.Listen(); err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() { _ = srv.Serve() }()
	t.Cleanup(func() { _ = srv.Stop(); store.CloseAll(); pool.CloseAll() })

	return srv.Address()
}

func seedExportTable(t *testing.T) {
	t.Helper()

	host, port := cassandraAddr(t)
	cluster := gocql.NewCluster(host)
	cluster.Port = port
	cluster.Consistency = gocql.One
	cluster.Timeout = 15 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		t.Skipf("cassandra not reachable at %s:%d: %v", host, port, err)
	}
	defer session.Close()

	mustExec := func(q string) {
		t.Helper()
		if err := session.Query(q).Exec(); err != nil {
			t.Fatalf("exec %q: %v", q, err)
		}
	}

	mustExec(fmt.Sprintf(`DROP KEYSPACE IF EXISTS %s`, exportKeyspace))
	mustExec(fmt.Sprintf(`CREATE KEYSPACE %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`, exportKeyspace))
	mustExec(fmt.Sprintf(`CREATE TABLE %s.export_items (id int, label text, PRIMARY KEY (id))`, exportKeyspace))

	batch := session.NewBatch(gocql.UnloggedBatch)
	for i := range exportRowCount {
		batch.Query(
			fmt.Sprintf(`INSERT INTO %s.export_items (id, label) VALUES (?, ?)`, exportKeyspace),
			i, fmt.Sprintf(`item-%d,"quoted",%s`, i, strings.Repeat("x", 4096)),
		)
		if (i+1)%10 == 0 {
			if err := session.ExecuteBatch(batch); err != nil {
				t.Fatalf("batch insert: %v", err)
			}
			batch = session.NewBatch(gocql.UnloggedBatch)
		}
	}
	if err := session.ExecuteBatch(batch); err != nil {
		t.Fatalf("batch insert: %v", err)
	}
}

func loginClient(t *testing.T, addr string) *client.Client {
	t.Helper()

	c, err := client.New(addr)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := c.Login(ctx, "it-export"); err != nil {
		t.Fatalf("login: %v", err)
	}
	return c
}

func findCSVRecord(records [][]string, id string) []string {
	for _, rec := range records[1:] {
		if len(rec) > 0 && rec[0] == id {
			return rec
		}
	}
	return nil
}

func TestExportRowsIntegration(t *testing.T) {
	seedExportTable(t)
	addr := startRealServer(t)
	c := loginClient(t, addr)

	ctx := context.Background()

	t.Run("csv export streams all rows", func(t *testing.T) {
		var sb strings.Builder
		var lastCount int64
		chunks := 0

		err := c.ExportRows(ctx, exportKeyspace, "export_items", "", pb.ExportFormat_EXPORT_FORMAT_CSV, 50, func(chunk *pb.ExportChunk) error {
			sb.Write(chunk.Data)
			lastCount = chunk.RowsExported
			chunks++
			return nil
		})
		if err != nil {
			t.Fatalf("export csv: %v", err)
		}
		if chunks < 2 {
			t.Errorf("expected multiple 512KB chunks for ~1MB payload, got %d", chunks)
		}
		if lastCount != exportRowCount {
			t.Errorf("rows_exported = %d, want %d", lastCount, exportRowCount)
		}

		records, err := csv.NewReader(strings.NewReader(sb.String())).ReadAll()
		if err != nil {
			t.Fatalf("parse csv: %v", err)
		}
		if len(records) != exportRowCount+1 {
			t.Fatalf("got %d csv records, want %d (header + rows)", len(records), exportRowCount+1)
		}
		if records[0][0] != "id" || records[0][1] != "label" {
			t.Errorf("unexpected csv header: %v", records[0])
		}

		row7 := findCSVRecord(records, "7")
		if row7 == nil {
			t.Fatal("row id=7 missing from export")
		}
		wantLabel := `item-7,"quoted"`
		if !strings.HasPrefix(row7[1], wantLabel) {
			t.Errorf("row 7 label = %q..., want prefix %q", row7[1][:30], wantLabel)
		}
	})

	t.Run("json export is newline-delimited objects", func(t *testing.T) {
		var sb strings.Builder
		err := c.ExportRows(ctx, exportKeyspace, "export_items", "", pb.ExportFormat_EXPORT_FORMAT_JSON, 60, func(chunk *pb.ExportChunk) error {
			sb.Write(chunk.Data)
			return nil
		})
		if err != nil {
			t.Fatalf("export json: %v", err)
		}

		lines := strings.Split(strings.TrimSuffix(sb.String(), "\n"), "\n")
		if len(lines) != exportRowCount {
			t.Fatalf("got %d json rows, want %d", len(lines), exportRowCount)
		}

		var row7 map[string]interface{}
		for _, line := range lines {
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(line), &obj); err != nil {
				t.Fatalf("unmarshal %q: %v", line[:40], err)
			}
			if obj["id"] == float64(7) {
				row7 = obj
			}
		}
		if row7 == nil {
			t.Fatal("row id=7 missing from json export")
		}
		if label, _ := row7["label"].(string); !strings.HasPrefix(label, `item-7,"quoted"`) {
			t.Errorf("row 7 label unexpected: %.40q", label)
		}
	})

	t.Run("filter export with where clause", func(t *testing.T) {
		var sb strings.Builder
		err := c.ExportRows(ctx, exportKeyspace, "export_items", "id = 5", pb.ExportFormat_EXPORT_FORMAT_CSV, 50, func(chunk *pb.ExportChunk) error {
			sb.Write(chunk.Data)
			return nil
		})
		if err != nil {
			t.Fatalf("filtered export: %v", err)
		}

		records, err := csv.NewReader(strings.NewReader(sb.String())).ReadAll()
		if err != nil {
			t.Fatalf("parse csv: %v", err)
		}
		if len(records) != 2 {
			t.Fatalf("got %d records, want header + 1 row", len(records))
		}
		if records[1][0] != "5" {
			t.Errorf("got id %q, want 5", records[1][0])
		}
	})

	t.Run("rejects invalid identifiers", func(t *testing.T) {
		err := c.ExportRows(ctx, "ks; DROP", "tbl", "", pb.ExportFormat_EXPORT_FORMAT_CSV, 10, func(*pb.ExportChunk) error { return nil })
		if err == nil || !strings.Contains(err.Error(), "invalid keyspace") {
			t.Fatalf("expected invalid keyspace error, got %v", err)
		}
	})

	t.Run("requires authentication on streaming rpc", func(t *testing.T) {
		anon, err := client.New(addr)
		if err != nil {
			t.Fatalf("client: %v", err)
		}
		defer anon.Close()

		err = anon.ExportRows(ctx, exportKeyspace, "export_items", "", pb.ExportFormat_EXPORT_FORMAT_CSV, 10, func(*pb.ExportChunk) error { return nil })
		if err == nil {
			t.Fatal("unauthenticated stream must fail")
		}
	})
}

func TestExecuteQueryIntegration(t *testing.T) {
	seedExportTable(t)
	addr := startRealServer(t)
	c := loginClient(t, addr)

	ctx := context.Background()

	t.Run("select all and page via cursor", func(t *testing.T) {
		resp, err := c.ExecuteQuery(ctx, fmt.Sprintf("SELECT id, label FROM %s.export_items", exportKeyspace), 100)
		if err != nil {
			t.Fatalf("execute query: %v", err)
		}
		if resp.TotalFetched != 100 || !resp.HasMore || resp.CursorId == "" {
			t.Fatalf("first page: total=%d hasMore=%v cursor=%q", resp.TotalFetched, resp.HasMore, resp.CursorId)
		}

		seen := map[float64]bool{}
		for _, row := range resp.Rows {
			if iv := cellID(row); iv >= 0 {
				seen[float64(iv)] = true
			}
		}

		pages := 1
		cursorID := resp.CursorId
		for cursorID != "" {
			pageResp, err := c.GetNextPage(ctx, cursorID)
			if err != nil {
				t.Fatalf("page %d: %v", pages+1, err)
			}
			for _, row := range pageResp.Rows {
				if iv := row.Cells["id"].GetIntVal(); iv != 0 {
					seen[float64(iv)] = true
				}
			}
			cursorID = pageResp.CursorId
			pages++
			if pages > 10 {
				t.Fatal("paging did not terminate")
			}
		}

		if len(seen) != exportRowCount {
			t.Fatalf("saw %d distinct ids across %d pages, want %d", len(seen), pages, exportRowCount)
		}
	})

	t.Run("rejects write statements", func(t *testing.T) {
		_, err := c.ExecuteQuery(ctx, fmt.Sprintf("DROP KEYSPACE %s", exportKeyspace), 10)
		if err == nil || !strings.Contains(err.Error(), "only SELECT") {
			t.Fatalf("expected SELECT-only rejection, got %v", err)
		}
	})

	t.Run("rejects multiple statements", func(t *testing.T) {
		_, err := c.ExecuteQuery(ctx, "SELECT a FROM b; SELECT c FROM d", 10)
		if err == nil || !strings.Contains(err.Error(), "multiple statements") {
			t.Fatalf("expected multiple statement rejection, got %v", err)
		}
	})
}

func cellID(row *pb.Row) int64 {
	cell, ok := row.Cells["id"]
	if !ok || cell == nil {
		return -1
	}
	if iv, ok := cell.Value.(*pb.CellValue_IntVal); ok {
		return iv.IntVal
	}
	return -1
}

func TestQueryHistoryIntegration(t *testing.T) {
	seedExportTable(t)
	addr := startRealServer(t)
	c := loginClient(t, addr)

	ctx := context.Background()

	t.Run("execute query records history", func(t *testing.T) {
		if _, err := c.ExecuteQuery(ctx, fmt.Sprintf("SELECT id FROM %s.export_items LIMIT 5", exportKeyspace), 10); err != nil {
			t.Fatalf("execute: %v", err)
		}
		if _, err := c.ExecuteQuery(ctx, fmt.Sprintf("SELECT count(*) FROM %s.export_items", exportKeyspace), 10); err != nil {
			t.Fatalf("execute: %v", err)
		}

		entries, err := c.ListQueryHistory(ctx, 10)
		if err != nil {
			t.Fatalf("list history: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("history entries = %d, want 2", len(entries))
		}
		if !strings.Contains(entries[0].Cql, "count(*)") {
			t.Errorf("most recent first: %q", entries[0].Cql)
		}

		if err := c.ClearQueryHistory(ctx); err != nil {
			t.Fatalf("clear: %v", err)
		}
		entries, err = c.ListQueryHistory(ctx, 10)
		if err != nil {
			t.Fatalf("re-list: %v", err)
		}
		if len(entries) != 0 {
			t.Fatalf("history after clear = %d, want 0", len(entries))
		}
	})

	t.Run("saved queries crud and validation", func(t *testing.T) {
		saved, err := c.SaveQuery(ctx, "top-items", fmt.Sprintf("SELECT * FROM %s.export_items LIMIT 10", exportKeyspace))
		if err != nil {
			t.Fatalf("save: %v", err)
		}
		if saved.Name != "top-items" {
			t.Errorf("saved name = %q", saved.Name)
		}

		list, err := c.ListSavedQueries(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(list) != 1 || list[0].Name != "top-items" {
			t.Fatalf("saved list = %+v", list)
		}

		if _, err := c.SaveQuery(ctx, "bad;name", "SELECT 1"); err == nil {
			t.Error("invalid name accepted")
		}

		if _, err := c.SaveQuery(ctx, "top-items", "SELECT id FROM x"); err != nil {
			t.Fatalf("upsert rejected: %v", err)
		}
		list, _ = c.ListSavedQueries(ctx)
		if len(list) != 1 || list[0].Cql != "SELECT id FROM x" {
			t.Fatalf("upsert result = %+v", list)
		}

		if _, err := c.ExecuteQuery(ctx, "SELECT id FROM missing_table_xyz", 5); err == nil {
			t.Log("note: query against missing table returned nil error")
		}

		if err := c.DeleteSavedQuery(ctx, "top-items"); err != nil {
			t.Fatalf("delete: %v", err)
		}
		if err := c.DeleteSavedQuery(ctx, "top-items"); err == nil {
			t.Error("deleting missing query must fail")
		}
	})
}

func TestTableStatsIntegration(t *testing.T) {
	seedExportTable(t)
	addr := startRealServer(t)
	c := loginClient(t, addr)

	stats, err := c.GetTableStats(context.Background(), exportKeyspace, "export_items")
	if err != nil {
		t.Fatalf("table stats: %v", err)
	}
	if stats.RowCount < exportRowCount && !stats.EstimateAvailable {
		t.Fatalf("row_count = %d, want >= %d (estimate_available=%v)", stats.RowCount, exportRowCount, stats.EstimateAvailable)
	}

	if _, err := c.GetTableStats(context.Background(), "bad;ks", "tbl"); err == nil {
		t.Error("invalid keyspace accepted")
	}
	if _, err := c.GetTableStats(context.Background(), exportKeyspace, ""); err == nil {
		t.Error("empty table accepted")
	}
}

func TestTypedCellsIntegration(t *testing.T) {
	host, port := cassandraAddr(t)
	cluster := gocql.NewCluster(host)
	cluster.Port = port
	cluster.Consistency = gocql.One
	cluster.Timeout = 15 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		t.Skipf("cassandra not reachable: %v", err)
	}

	ks := "kassie_types_it"
	must := func(q string) {
		t.Helper()
		if err := session.Query(q).Exec(); err != nil {
			t.Fatalf("exec %q: %v", q, err)
		}
	}
	must(fmt.Sprintf(`DROP KEYSPACE IF EXISTS %s`, ks))
	must(fmt.Sprintf(`CREATE KEYSPACE %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`, ks))
	must(fmt.Sprintf(`CREATE TABLE %s.rich (id int PRIMARY KEY, payload blob, attrs map<text,int>, tags list<text>, at timestamp, uid uuid)`, ks))
	must(fmt.Sprintf(`INSERT INTO %s.rich (id, payload, attrs, tags, at, uid) VALUES (1, 0xdeadbeef, {'a': 1}, ['x','y'], '2026-08-26 12:00:00+0000', 550e8400-e29b-41d4-a716-446655440000)`, ks))
	session.Close()

	addr := startRealServer(t)
	c := loginClient(t, addr)

	resp, err := c.QueryRows(context.Background(), ks, "rich", 10)
	if err != nil {
		t.Fatalf("query rows: %v", err)
	}
	if len(resp.Rows) != 1 {
		t.Fatalf("rows = %d", len(resp.Rows))
	}

	cells := resp.Rows[0].Cells

	if b := cells["payload"].GetBytesVal(); len(b) != 4 || b[0] != 0xDE {
		t.Errorf("blob roundtrip failed: %x", b)
	}
	if cells["payload"].CqlType != "blob" {
		t.Errorf("blob cql_type = %q", cells["payload"].CqlType)
	}

	if got := cells["attrs"].GetStringVal(); got != `{"a":1}` {
		t.Errorf("map json = %q", got)
	}
	if cells["attrs"].CqlType != "map<varchar, int>" {
		t.Errorf("map cql_type = %q", cells["attrs"].CqlType)
	}

	if got := cells["tags"].GetStringVal(); got != `["x","y"]` {
		t.Errorf("list json = %q", got)
	}

	if got := cells["at"].GetStringVal(); !strings.HasPrefix(got, "2026-08-26T12:00:00") {
		t.Errorf("timestamp = %q", got)
	}

	if got := cells["uid"].GetStringVal(); got != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("uuid = %q", got)
	}

	first, err := c.QueryRows(context.Background(), ks, "rich", 1)
	if err != nil {
		t.Fatalf("page one: %v", err)
	}
	if !first.HasMore {
		t.Skip("table too small to exercise typed GetNextPage")
	}
	next, err := c.GetNextPage(context.Background(), first.CursorId)
	if err != nil {
		t.Fatalf("typed next page: %v", err)
	}
	if len(next.Rows) > 0 {
		if cell := next.Rows[0].Cells["payload"]; cell != nil && cell.CqlType != "blob" {
			t.Errorf("GetNextPage lost type info: %q", cell.CqlType)
		}
	}
}

func TestMetricsIntegration(t *testing.T) {
	seedExportTable(t)
	addr := startRealServer(t)
	c := loginClient(t, addr)

	ctx := context.Background()

	if _, err := c.ExecuteQuery(ctx, fmt.Sprintf("SELECT id FROM %s.export_items LIMIT 5", exportKeyspace), 2); err != nil {
		t.Fatalf("setup query: %v", err)
	}

	metrics, err := c.GetMetrics(ctx)
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if metrics.ActiveSessions < 1 {
		t.Errorf("active_sessions = %d, want >= 1", metrics.ActiveSessions)
	}
	if metrics.ActiveCursors < 1 {
		t.Errorf("active_cursors = %d, want >= 1 (query with has_more creates cursor)", metrics.ActiveCursors)
	}

	anon, err := client.New(addr)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer anon.Close()
	if _, err := anon.GetMetrics(ctx); err == nil {
		t.Error("unauthenticated metrics must fail")
	}
}
