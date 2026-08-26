//go:build integration

package grpc_test

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
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
		Config: provider,
		Pool:   pool,
		Store:  store,
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
