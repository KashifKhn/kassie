package cache

import (
	"sync"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
)

type schemaEntry struct {
	schema    *pb.TableSchema
	timestamp time.Time
}

type SchemaCache struct {
	mu      sync.RWMutex
	entries map[string]schemaEntry
	ttl     time.Duration
	hits    int
	misses  int
}

func NewSchemaCache(ttl time.Duration) *SchemaCache {
	return &SchemaCache{
		entries: make(map[string]schemaEntry),
		ttl:     ttl,
	}
}

func (c *SchemaCache) Get(keyspace, table string) (*pb.TableSchema, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := keyspace + "." + table
	entry, exists := c.entries[key]

	if !exists {
		c.mu.RUnlock()
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		c.mu.RLock()
		return nil, false
	}

	if time.Since(entry.timestamp) > c.ttl {
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.entries, key)
		c.misses++
		c.mu.Unlock()
		c.mu.RLock()
		return nil, false
	}

	c.mu.RUnlock()
	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
	c.mu.RLock()
	return entry.schema, true
}

func (c *SchemaCache) Set(keyspace, table string, schema *pb.TableSchema) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := keyspace + "." + table
	c.entries[key] = schemaEntry{
		schema:    schema,
		timestamp: time.Now(),
	}
}

func (c *SchemaCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]schemaEntry)
	c.hits = 0
	c.misses = 0
}

func (c *SchemaCache) Invalidate(keyspace, table string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := keyspace + "." + table
	delete(c.entries, key)
}

func (c *SchemaCache) Stats() (hits, misses, size int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.hits, c.misses, len(c.entries)
}
