# Client-Server Model

Kassie's architecture separates concerns between clients (TUI and Web) and a central server that manages all database interactions.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    KASSIE UNIFIED BINARY                        │
├─────────────────┬───────────────────┬──────────────────────────┤
│  kassie tui     │  kassie web       │  kassie server           │
│  (TUI mode)     │  (Web mode)       │  (Daemon mode)           │
└────────┬────────┴─────────┬─────────┴──────────┬───────────────┘
         │                  │                    │
         │  Embedded        │  Embedded          │  Standalone
         │  Server          │  Server            │  Server
         │                  │  + HTTP            │  + HTTP
         └──────────┬───────┴────────────────────┘
                    │
        ┌───────────▼────────────────────────────────────┐
        │         KASSIE SERVER CORE                     │
        │  ├── gRPC API Layer                            │
        │  ├── HTTP Gateway (grpc-gateway)               │
        │  ├── Service Layer                             │
        │  ├── Connection Pool Manager                   │
        │  └── Session State Store                       │
        └───────────┬────────────────────────────────────┘
                    │
        ┌───────────▼────────────────────────────────────┐
        │      Cassandra / ScyllaDB Cluster              │
        └────────────────────────────────────────────────┘
```

## Deployment Modes

### Embedded Mode (TUI / Web)

When you run `kassie tui` or `kassie web`, the server starts as a **background goroutine** within the same process:

- Binds to `localhost` only (not network-accessible)
- Shares the process lifecycle with the client
- Minimal overhead — no separate process needed
- Auto-generated JWT secret (no configuration required)

```bash
# Server starts automatically in the background
kassie tui
kassie web
```

### Standalone Mode (Server)

When you run `kassie server`, it starts as an **independent daemon**:

- Binds to configurable network interface (default `0.0.0.0`)
- Supports multiple concurrent clients
- Independent lifecycle — survives client disconnects
- Requires explicit JWT secret for security

```bash
# Start standalone server
kassie server --grpc-port 50051 --http-port 8080

# Connect TUI to remote server
kassie tui --server remote-host:50051
```

## Server Core Components

### gRPC Server

The primary API layer using Protocol Buffers for type-safe communication:

- Three service definitions: `SessionService`, `SchemaService`, `DataService`
- Unary RPCs for all operations
- Auth interceptor validates JWT on every request
- Reflection enabled for debugging with tools like `grpcurl`

### HTTP Gateway (grpc-gateway)

Auto-generated REST API that translates HTTP/JSON to gRPC:

- All endpoints under `/api/v1/`
- Serves the Web UI static files at `/`
- SPA fallback — non-API routes serve `index.html`
- CORS configured for development

### Service Layer

Business logic separated from transport:

| Service | Responsibility |
|---------|---------------|
| `AuthService` | JWT generation, validation, refresh |
| `SessionService` | Profile loading, session lifecycle |
| `SchemaService` | Keyspace/table introspection via system tables |
| `DataService` | Query execution, pagination, filtering |

### Connection Pool Manager

Per-profile connection pools to Cassandra/ScyllaDB:

- 5 connections per profile (gocql default)
- Automatic reconnection with exponential backoff
- Health checks every 30 seconds
- SSL/TLS support

### Session State Store

In-memory state management:

- Maps `session_id` to session data (profile, connection, cursors)
- Cursor store for pagination state tokens
- Auto-cleanup on session expiry

## Communication Patterns

### TUI Client → Server

```
TUI ──(gRPC)──> Server ──(gocql)──> Cassandra
```

Direct gRPC for maximum performance (~1-2ms latency locally).

### Web Client → Server

```
Browser ──(HTTP/JSON)──> Gateway ──(gRPC)──> Server ──(gocql)──> Cassandra
```

REST via grpc-gateway (~3-5ms latency locally). JSON payloads.

### Request Lifecycle

1. Client sends request (gRPC or HTTP)
2. Auth interceptor validates JWT token
3. Service layer processes business logic
4. Database layer executes CQL query
5. Response serialized and returned to client

## Web Asset Embedding

The Web UI ships inside the Go binary:

1. `make web` builds React app → `web/dist/`
2. `make embed-web` copies to `internal/server/web/dist/`
3. Go's `embed.FS` packages assets into the binary
4. Gateway serves static files at `/`, API at `/api/`

```go
//go:embed dist/*
var webAssets embed.FS
```

In development, the `--web-root` flag serves from the filesystem instead of `embed.FS`, enabling Vite hot reload.
