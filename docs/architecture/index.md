# Architecture

Understanding Kassie's design and implementation.

## Overview

Kassie follows a client-server architecture with these key components:

- **Server Core**: gRPC server with HTTP gateway
- **TUI Client**: Bubbletea-based terminal interface
- **Web Client**: React-based browser interface (in development)
- **Shared Client SDK**: Common gRPC client wrapper

## Key Design Principles

1. **Dual-Client Equality**: TUI and Web as first-class citizens
2. **Single Binary**: All-in-one executable with embedded assets
3. **Type-Safe Communication**: gRPC with auto-generated clients
4. **Read-Safety First**: Optimized for browsing and observing data
5. **Embedded Server**: Server runs within client process or standalone

## Deployment Modes

### Embedded Mode
Server starts as background goroutine within TUI/Web client process.

### Standalone Mode
Server runs independently, accepting remote client connections.

---

## Performance Characteristics

### Memory Usage

**TUI Client (Embedded)**:
- Base memory: ~15-25 MB (includes embedded server)
- Per-connection: +5-10 MB per active Cassandra connection
- Data display: Minimal (streaming, no caching)
- Large result sets: Pagination prevents unbounded memory growth

**Standalone Server**:
- Base memory: ~20-30 MB
- Per-session: +5-10 MB per active client session
- Connection pool: ~10-15 MB per database profile
- Cursors: ~1-5 MB per active cursor (expires after 30 min)

**Memory Scaling**:
```
Base + (Connections × 10MB) + (Active Cursors × 5MB)
```

### Connection Pooling

Kassie manages Cassandra connections automatically:

| Setting | Value | Configuration |
|---------|-------|---------------|
| Pool size per profile | 5 | Automatic (gocql default) |
| Max connections | 2 per host | Automatic (gocql default) |
| Health check interval | 30 seconds | Automatic |
| Idle timeout | 10 minutes | Automatic |
| Reconnect backoff | Exponential (1s → 30s) | Automatic |

**Connection lifecycle**:
1. Client logs in with profile → Connection pool created
2. Idle connections reused across requests
3. Failed connections trigger automatic reconnect
4. Logout/session timeout → Pool closed

### Query Timeouts

| Operation | Default Timeout | Configurable |
|-----------|-----------------|--------------|
| Schema queries | 10 seconds | Via `defaults.timeout_ms` |
| Data queries | 10 seconds | Via `defaults.timeout_ms` |
| Login/session | 10 seconds | Hardcoded |
| Cursor fetch | Inherits query timeout | Via `defaults.timeout_ms` |

**Timeout configuration** (`config.json`):
```json
{
  "defaults": {
    "timeout_ms": 10000
  }
}
```

**Range**: 100ms - 300,000ms (5 minutes)

### Pagination Performance

**Default page size**: 100 rows

| Page Size | Initial Load | Memory Impact | Network | Use Case |
|-----------|--------------|---------------|---------|----------|
| 50 | Fast (~100ms) | Low | Low | Quick browsing |
| 100 | Good (~200ms) | Medium | Medium | **Recommended** |
| 500 | Slower (~1s) | High | High | Data export |
| 1000+ | Slow (2s+) | Very High | Very High | Bulk operations |

**Performance tips**:
- Smaller page sizes = faster initial response
- Larger page sizes = fewer round trips
- Use filters to reduce total dataset
- Cursors are stateful (held in server memory)

**Cursor expiration**: 30 minutes of inactivity

### gRPC vs HTTP Performance

| Metric | gRPC | HTTP/REST | Notes |
|--------|------|-----------|-------|
| Latency | ~1-2ms | ~3-5ms | Local embedded server |
| Payload size | Smaller (protobuf) | Larger (JSON) | ~30-40% difference |
| Throughput | Higher | Lower | Binary encoding advantage |
| CPU | Lower | Higher | JSON parsing overhead |

**Recommendation**: Use gRPC for custom clients, HTTP/REST for quick scripts and testing.

### Scalability Limits

**Tested configurations**:
| Scenario | Performance | Notes |
|----------|-------------|-------|
| 100M+ row table | ✅ Good | With pagination and filters |
| 1000+ columns (wide row) | ⚠️ Slow | TUI rendering bottleneck |
| 10+ concurrent clients | ✅ Good | Standalone server mode |
| 50+ concurrent clients | ⚠️ Untested | May need tuning |
| 1GB+ blob columns | ❌ Poor | Not optimized for large blobs |

**Known bottlenecks**:
- TUI rendering with >500 columns visible
- Inspector panel with >10,000 lines
- HTTP gateway JSON marshaling for wide rows

### Network Considerations

**Bandwidth usage** (approximate):
- Schema introspection: <100 KB per keyspace
- 100-row page fetch: 10-500 KB (depends on column width)
- gRPC overhead: ~50 bytes per request
- JWT token: ~500 bytes per request (HTTP only)

**Latency impact**:
- Local cluster: <5ms query latency
- Same datacenter: 5-20ms
- Cross-region: 50-200ms
- Pagination helps with high-latency connections

---

## Security

### Authentication

**JWT-based authentication**:
- Access tokens: 1 hour expiration
- Refresh tokens: 24 hour expiration
- HMAC-SHA256 signing

**Token storage**:
- TUI: In-memory only (not persisted)
- Server: Stateless (no token storage)

### Embedded vs Standalone Security

| Mode | JWT Secret | Threat Model |
|------|-----------|--------------|
| Embedded (TUI/Web) | Auto-generated | Local process, low risk |
| Standalone Server | User-provided | Network-exposed, high risk |

**Production recommendation**: Always set `KASSIE_JWT_SECRET` for standalone servers.

### gRPC Connection Security

::: warning Coming Soon
TLS/SSL support for gRPC connections is planned but not yet implemented.

**Current status**: Plain-text gRPC connections only  
**Workaround**: Use SSH tunneling or VPN for remote access  
**Planned**: Full TLS support with client certificates
:::

**Planned features**:
- [ ] TLS/SSL for gRPC server
- [ ] Mutual TLS (mTLS) with client certificates
- [ ] Certificate validation
- [ ] Token revocation API

**SSH Tunnel workaround** (current recommendation):
```bash
# On local machine
ssh -L 50051:localhost:50051 remote-server

# Connect via tunnel
kassie tui --server localhost:50051
```

### Database Connection Security

SSL/TLS to Cassandra/ScyllaDB:
- ✅ Supported via config
- ✅ Client certificates
- ✅ CA validation
- ⚠️ Can disable verification (insecure)

**Example secure config**:
```json
{
  "profiles": [{
    "ssl": {
      "enabled": true,
      "cert_path": "${HOME}/.cassandra/client.crt",
      "key_path": "${HOME}/.cassandra/client.key",
      "ca_path": "${HOME}/.cassandra/ca.crt",
      "insecure_skip_verify": false
    }
  }]
}
```

---

## Learn More

- [Client-Server Model](/architecture/client-server) - Detailed architecture *(Coming Soon)*
- [Authentication](/architecture/authentication) - Auth system design *(Coming Soon)*
- [State Management](/architecture/state-management) - State handling *(Coming Soon)*
- [Protocol Design](/architecture/protocol) - gRPC protocol details *(Coming Soon)*

For the complete architecture document, see [n_docs/ARCHITECTURE.md](https://github.com/KashifKhn/kassie/blob/main/n_docs/ARCHITECTURE.md) in the repository.
