package cache

import (
	"testing"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
)

func TestSchemaCache_SetAndGet(t *testing.T) {
	cache := NewSchemaCache(5 * time.Minute)

	schema := &pb.TableSchema{
		Columns: []*pb.Column{
			{Name: "id", Type: "uuid"},
			{Name: "name", Type: "text"},
		},
	}

	cache.Set("keyspace1", "table1", schema)

	retrieved, found := cache.Get("keyspace1", "table1")
	if !found {
		t.Fatal("expected schema to be found")
	}

	if len(retrieved.Columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(retrieved.Columns))
	}
}

func TestSchemaCache_Miss(t *testing.T) {
	cache := NewSchemaCache(5 * time.Minute)

	_, found := cache.Get("nonexistent", "table")
	if found {
		t.Error("expected cache miss for nonexistent key")
	}

	hits, misses, _ := cache.Stats()
	if hits != 0 || misses != 1 {
		t.Errorf("expected 0 hits and 1 miss, got %d hits and %d misses", hits, misses)
	}
}

func TestSchemaCache_TTL(t *testing.T) {
	cache := NewSchemaCache(100 * time.Millisecond)

	schema := &pb.TableSchema{
		Columns: []*pb.Column{{Name: "id", Type: "uuid"}},
	}

	cache.Set("ks", "tb", schema)

	_, found := cache.Get("ks", "tb")
	if !found {
		t.Fatal("expected schema to be found immediately")
	}

	time.Sleep(150 * time.Millisecond)

	_, found = cache.Get("ks", "tb")
	if found {
		t.Error("expected schema to be expired after TTL")
	}
}

func TestSchemaCache_Clear(t *testing.T) {
	cache := NewSchemaCache(5 * time.Minute)

	schema := &pb.TableSchema{Columns: []*pb.Column{{Name: "id", Type: "uuid"}}}

	cache.Set("ks1", "tb1", schema)
	cache.Set("ks2", "tb2", schema)

	_, _, size := cache.Stats()
	if size != 2 {
		t.Errorf("expected cache size 2, got %d", size)
	}

	cache.Clear()

	hits, misses, size := cache.Stats()
	if size != 0 || hits != 0 || misses != 0 {
		t.Errorf("expected empty cache after clear, got size=%d hits=%d misses=%d", size, hits, misses)
	}
}

func TestSchemaCache_Invalidate(t *testing.T) {
	cache := NewSchemaCache(5 * time.Minute)

	schema := &pb.TableSchema{Columns: []*pb.Column{{Name: "id", Type: "uuid"}}}

	cache.Set("ks1", "tb1", schema)
	cache.Set("ks1", "tb2", schema)

	cache.Invalidate("ks1", "tb1")

	_, found := cache.Get("ks1", "tb1")
	if found {
		t.Error("expected tb1 to be invalidated")
	}

	_, found = cache.Get("ks1", "tb2")
	if !found {
		t.Error("expected tb2 to still exist")
	}
}

func TestSchemaCache_Stats(t *testing.T) {
	cache := NewSchemaCache(5 * time.Minute)

	schema := &pb.TableSchema{Columns: []*pb.Column{{Name: "id", Type: "uuid"}}}

	cache.Set("ks", "tb", schema)

	cache.Get("ks", "tb")
	cache.Get("ks", "tb")
	cache.Get("ks", "nonexistent")

	hits, misses, size := cache.Stats()

	if hits != 2 {
		t.Errorf("expected 2 hits, got %d", hits)
	}

	if misses != 1 {
		t.Errorf("expected 1 miss, got %d", misses)
	}

	if size != 1 {
		t.Errorf("expected cache size 1, got %d", size)
	}
}

func TestSchemaCache_ConcurrentAccess(t *testing.T) {
	cache := NewSchemaCache(5 * time.Minute)

	schema := &pb.TableSchema{Columns: []*pb.Column{{Name: "id", Type: "uuid"}}}

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				cache.Set("ks", "tb", schema)
				cache.Get("ks", "tb")
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	hits, misses, size := cache.Stats()
	if size != 1 {
		t.Errorf("expected cache size 1 after concurrent writes, got %d", size)
	}

	if hits+misses != 1000 {
		t.Errorf("expected 1000 total operations, got %d", hits+misses)
	}
}

func TestSchemaCache_OverwriteEntry(t *testing.T) {
	cache := NewSchemaCache(5 * time.Minute)

	schema1 := &pb.TableSchema{
		Columns: []*pb.Column{{Name: "id", Type: "uuid"}},
	}

	schema2 := &pb.TableSchema{
		Columns: []*pb.Column{
			{Name: "id", Type: "uuid"},
			{Name: "name", Type: "text"},
		},
	}

	cache.Set("ks", "tb", schema1)
	cache.Set("ks", "tb", schema2)

	retrieved, found := cache.Get("ks", "tb")
	if !found {
		t.Fatal("expected schema to be found")
	}

	if len(retrieved.Columns) != 2 {
		t.Errorf("expected overwritten schema with 2 columns, got %d", len(retrieved.Columns))
	}
}
