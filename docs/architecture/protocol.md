# Protocol Design

Kassie uses Protocol Buffers (protobuf) for API definitions, with gRPC as the primary transport and grpc-gateway providing REST/JSON endpoints for web clients.

## Proto File Structure

All definitions live in `api/proto/`:

```
api/proto/
├── common.proto    # Shared types (Column, CellValue, Error, ViewState)
├── session.proto   # SessionService (Login, Refresh, Logout, GetProfiles)
├── schema.proto    # SchemaService (ListKeyspaces, ListTables, GetTableSchema)
└── data.proto      # DataService (QueryRows, GetNextPage, FilterRows)
```

## Service Definitions

### SessionService

Manages authentication and profile access.

| RPC | HTTP Mapping | Description |
|-----|-------------|-------------|
| `Login` | `POST /api/v1/session/login` | Authenticate with a profile |
| `Refresh` | `POST /api/v1/session/refresh` | Obtain new access token |
| `Logout` | `POST /api/v1/session/logout` | End session |
| `GetProfiles` | `GET /api/v1/profiles` | List available profiles |

**Login Request/Response:**
```json
// POST /api/v1/session/login
// Request
{ "profile": "local" }

// Response
{
  "access_token": "eyJhbG...",
  "refresh_token": "eyJhbG...",
  "expires_at": 1700000000,
  "profile": {
    "name": "local",
    "hosts": ["127.0.0.1"],
    "port": 9042,
    "keyspace": "",
    "ssl_enabled": false
  }
}
```

### SchemaService

Introspects database schema (keyspaces, tables, columns).

| RPC | HTTP Mapping | Description |
|-----|-------------|-------------|
| `ListKeyspaces` | `GET /api/v1/schema/keyspaces` | List all keyspaces |
| `ListTables` | `GET /api/v1/schema/keyspaces/{keyspace}/tables` | List tables in a keyspace |
| `GetTableSchema` | `GET /api/v1/schema/keyspaces/{keyspace}/tables/{table}` | Get column definitions |

**Table Schema Response:**
```json
// GET /api/v1/schema/keyspaces/app_data/tables/users
{
  "schema": {
    "keyspace": "app_data",
    "table": "users",
    "columns": [
      { "name": "id", "type": "uuid", "is_partition_key": true, "is_clustering_key": false, "position": 0 },
      { "name": "email", "type": "text", "is_partition_key": false, "is_clustering_key": false, "position": 1 },
      { "name": "created_at", "type": "timestamp", "is_partition_key": false, "is_clustering_key": true, "position": 0 }
    ],
    "partition_keys": ["id"],
    "clustering_keys": ["created_at"]
  }
}
```

### DataService

Queries and filters table data with cursor-based pagination.

| RPC | HTTP Mapping | Description |
|-----|-------------|-------------|
| `QueryRows` | `POST /api/v1/data/query` | Fetch initial rows |
| `GetNextPage` | `POST /api/v1/data/next` | Fetch next page via cursor |
| `FilterRows` | `POST /api/v1/data/filter` | Query with WHERE clause |

**Query and Pagination:**
```json
// POST /api/v1/data/query
// Request
{ "keyspace": "app_data", "table": "users", "page_size": 100 }

// Response
{
  "rows": [
    { "cells": { "id": { "string_val": "550e8400-..." }, "email": { "string_val": "user@example.com" } } }
  ],
  "cursor_id": "abc-123",
  "has_more": true,
  "total_fetched": 100
}

// POST /api/v1/data/next
// Request
{ "cursor_id": "abc-123" }

// Response
{
  "rows": [...],
  "cursor_id": "abc-456",
  "has_more": false
}
```

**Filtering:**
```json
// POST /api/v1/data/filter
// Request
{
  "keyspace": "app_data",
  "table": "users",
  "where_clause": "id = '550e8400-e29b-41d4-a716-446655440000'",
  "page_size": 100
}
```

## Common Message Types

### Column

Describes a table column with key information:

```protobuf
message Column {
  string name = 1;
  string type = 2;
  bool is_partition_key = 3;
  bool is_clustering_key = 4;
  int32 position = 5;
}
```

### CellValue

Discriminated union for typed cell values:

```protobuf
message CellValue {
  oneof value {
    string string_val = 1;
    int64 int_val = 2;
    double double_val = 3;
    bool bool_val = 4;
    bytes bytes_val = 5;
  }
  bool is_null = 6;
}
```

All Cassandra types are mapped to one of these protobuf types:

| Cassandra Type | Protobuf Field | Notes |
|---------------|----------------|-------|
| text, varchar, ascii | `string_val` | |
| uuid, timeuuid | `string_val` | Formatted as string |
| int, bigint, varint, counter | `int_val` | |
| float, double, decimal | `double_val` | |
| boolean | `bool_val` | |
| blob | `bytes_val` | |
| timestamp | `string_val` | ISO 8601 format |
| list, set, map, tuple, UDT | `string_val` | JSON-serialized |
| null | — | `is_null = true` |

### Row

A map of column names to cell values:

```protobuf
message Row {
  map<string, CellValue> cells = 1;
}
```

### Error

Structured error with code and details:

```protobuf
message Error {
  string code = 1;
  string message = 2;
  map<string, string> details = 3;
}
```

## Error Codes

| Code | gRPC Status | HTTP Status | Meaning |
|------|------------|-------------|---------|
| `AUTH_REQUIRED` | Unauthenticated | 401 | No token provided |
| `AUTH_INVALID` | Unauthenticated | 401 | Token invalid or expired |
| `AUTH_FORBIDDEN` | PermissionDenied | 403 | Insufficient permissions |
| `PROFILE_NOT_FOUND` | NotFound | 404 | Profile doesn't exist in config |
| `CONNECTION_FAILED` | Unavailable | 503 | Cannot connect to database |
| `QUERY_ERROR` | Internal | 500 | CQL execution failed |
| `INVALID_FILTER` | InvalidArgument | 400 | WHERE clause syntax error |
| `CURSOR_EXPIRED` | NotFound | 404 | Pagination cursor no longer valid |
| `INTERNAL` | Internal | 500 | Unexpected server error |

## Pagination Strategy

Kassie uses **cursor-based pagination** built on Scylla's native paging state:

1. Client sends `QueryRows` with `page_size`
2. Server executes CQL with `PageSize` set
3. Server stores Scylla's `paging_state` token as a cursor
4. Client receives `cursor_id` and `has_more` flag
5. Client sends `GetNextPage` with `cursor_id` for next page
6. Process repeats until `has_more` is `false`

**Advantages over offset-based pagination:**
- Consistent results even with concurrent writes
- No `ALLOW FILTERING` needed
- Efficient — Scylla resumes from exact position
- No duplicate or missing rows between pages

**Cursor expiration:** 30 minutes of inactivity. Expired cursors return `CURSOR_EXPIRED`.

## Code Generation

Generated code is placed in `api/gen/`:

```bash
# Generate all protobuf code
make proto

# This runs scripts/gen-proto.sh which generates:
# - api/gen/go/    → Go server stubs and message types
# - gRPC service implementations
# - grpc-gateway reverse proxy
```

The generated Go code is imported by the server:

```go
import kassiev1 "github.com/KashifKhn/kassie/api/gen/go"
```
