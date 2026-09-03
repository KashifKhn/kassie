//go:build integration

package db

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gocql/gocql"
)

func integrationSession(t *testing.T) *gocql.Session {
	t.Helper()

	host := getenvDefault("KASSIE_TEST_CASSANDRA_HOST", "127.0.0.1")
	port := getenvDefault("KASSIE_TEST_CASSANDRA_PORT", "9042")

	cluster := gocql.NewCluster(host)
	cluster.Port = mustPort(t, port)
	cluster.Consistency = gocql.One
	cluster.Timeout = 15 * time.Second
	cluster.ConnectTimeout = 30 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		t.Skipf("cassandra not reachable at %s:%s: %v", host, port, err)
	}
	return session
}

func setupPagingTable(t *testing.T, session *gocql.Session) {
	t.Helper()

	err := session.Query(
		`CREATE KEYSPACE IF NOT EXISTS kassie_it WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`,
	).Exec()
	if err != nil {
		t.Fatalf("create keyspace: %v", err)
	}

	err = session.Query(`DROP TABLE IF EXISTS kassie_it.gps_tracking_points`).Exec()
	if err != nil {
		t.Fatalf("drop table: %v", err)
	}

	err = session.Query(
		`CREATE TABLE kassie_it.gps_tracking_points (
			device_id text,
			timestamp bigint,
			payload text,
			PRIMARY KEY (device_id, timestamp)
		)`,
	).Exec()
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	batch := session.NewBatch(gocql.UnloggedBatch)
	const rows = 1000
	widePayload := strings.Repeat("x", 10240)
	for i := range rows {
		batch.Query(
			`INSERT INTO kassie_it.gps_tracking_points (device_id, timestamp, payload) VALUES (?, ?, ?)`,
			"device-1", int64(i), widePayload,
		)
		if (i+1)%100 == 0 {
			if err := session.ExecuteBatch(batch); err != nil {
				t.Fatalf("insert batch %d: %v", i/100, err)
			}
			batch = session.NewBatch(gocql.UnloggedBatch)
		}
	}

	var count int64
	if err := session.Query(`SELECT COUNT(*) FROM kassie_it.gps_tracking_points`).Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != rows {
		t.Fatalf("expected %d seeded rows, got %d", rows, count)
	}
}

func TestFetchWithPagingIntegration(t *testing.T) {
	session := integrationSession(t)
	defer session.Close()

	setupPagingTable(t, session)
	s := NewSession(session)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	const pageSize = 100
	stmt := `SELECT device_id, timestamp, payload FROM kassie_it.gps_tracking_points`

	rows, pageState, err := s.FetchWithPaging(ctx, stmt, pageSize, nil)
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	if len(rows) != pageSize {
		t.Fatalf("first page returned %d rows, want exactly %d (drain-all regression)", len(rows), pageSize)
	}
	if len(pageState) == 0 {
		t.Fatal("first page returned empty page state but more rows exist")
	}

	seen := make(map[int64]bool, pageSize)
	for _, row := range rows {
		ts, ok := row["timestamp"].(int64)
		if !ok {
			t.Fatalf("row timestamp has type %T, want int64", row["timestamp"])
		}
		if seen[ts] {
			t.Fatalf("duplicate timestamp %d across pages", ts)
		}
		seen[ts] = true
	}

	pages := 1
	for len(pageState) > 0 {
		rows, pageState, err = s.FetchWithPaging(ctx, stmt, pageSize, pageState)
		if err != nil {
			t.Fatalf("page %d: %v", pages+1, err)
		}
		pages++

		for _, row := range rows {
			ts := row["timestamp"].(int64)
			if seen[ts] {
				t.Fatalf("duplicate timestamp %d across pages", ts)
			}
			seen[ts] = true
		}

		if len(rows) < pageSize && len(pageState) > 0 {
			t.Fatalf("short page of %d rows still reported more data", len(rows))
		}
		if pages > 50 {
			t.Fatal("paging did not terminate")
		}
	}

	if len(seen) != 1000 {
		t.Fatalf("walked all pages but saw %d distinct rows, want 1000", len(seen))
	}

	fmt.Printf("paged through %d pages, %d distinct rows total\n", pages, len(seen))
}

func getenvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustPort(t *testing.T, port string) int {
	t.Helper()
	var p int
	if _, err := fmt.Sscanf(port, "%d", &p); err != nil || p < 1 || p > 65535 {
		t.Fatalf("invalid port %q", port)
	}
	return p
}
