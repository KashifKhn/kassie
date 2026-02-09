# API Reference

Kassie exposes both gRPC and REST APIs for programmatic access. The gRPC services are automatically exposed as REST endpoints via grpc-gateway.

## Base URLs

| Protocol | Default Address | Description |
|----------|----------------|-------------|
| gRPC | `localhost:50051` | Binary protocol for high-performance clients |
| HTTP/REST | `http://localhost:8080` | JSON-based REST API (grpc-gateway) |

## Authentication

All API endpoints (except `/api/v1/profiles` and `/api/v1/session/login`) require authentication via JWT tokens.

### Obtaining a Token

```bash
# Login to get access token
curl -X POST http://localhost:8080/api/v1/session/login \
  -H "Content-Type: application/json" \
  -d '{"profile": "local"}'
```

**Response:**
```json
{
  "access_token": "eyJhbGc...token",
  "refresh_token": "eyJhbGc...refresh",
  "expires_at": 1707500000,
  "profile": {
    "name": "local",
    "hosts": ["127.0.0.1"],
    "port": 9042,
    "keyspace": "system",
    "ssl_enabled": false
  }
}
```

### Using the Token

Include the access token in the `Authorization` header:

```bash
curl -H "Authorization: Bearer eyJhbGc...token" \
  http://localhost:8080/api/v1/schema/keyspaces
```

---

## SessionService

Handles authentication, profile management, and session lifecycle.

### Login

**POST** `/api/v1/session/login`

Authenticate with a configured profile and obtain access/refresh tokens.

**Request:**
```json
{
  "profile": "local"
}
```

**Response:**
```json
{
  "access_token": "string",
  "refresh_token": "string",
  "expires_at": 1707500000,
  "profile": {
    "name": "local",
    "hosts": ["127.0.0.1"],
    "port": 9042,
    "keyspace": "system",
    "ssl_enabled": false
  }
}
```

**Status Codes:**
- `200`: Success
- `400`: Invalid profile name
- `401`: Authentication failed
- `500`: Server error

---

### Refresh Token

**POST** `/api/v1/session/refresh`

Refresh an expired access token using the refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGc...refresh"
}
```

**Response:**
```json
{
  "access_token": "string",
  "expires_at": 1707500000
}
```

**Status Codes:**
- `200`: Success
- `401`: Invalid or expired refresh token
- `500`: Server error

---

### Logout

**POST** `/api/v1/session/logout`

Invalidate the current session and tokens.

**Request:**
```json
{}
```

**Response:**
```json
{}
```

**Requires:** Authorization header

**Status Codes:**
- `200`: Success
- `401`: Unauthorized

---

### Get Profiles

**GET** `/api/v1/profiles`

List all configured connection profiles.

**Request:** None

**Response:**
```json
{
  "profiles": [
    {
      "name": "local",
      "hosts": ["127.0.0.1"],
      "port": 9042,
      "keyspace": "system",
      "ssl_enabled": false
    },
    {
      "name": "production",
      "hosts": ["prod-1.example.com", "prod-2.example.com"],
      "port": 9042,
      "keyspace": "app_data",
      "ssl_enabled": true
    }
  ]
}
```

**Status Codes:**
- `200`: Success
- `500`: Server error

**Note:** This endpoint does not require authentication.

---

## SchemaService

Provides schema introspection for keyspaces, tables, and columns.

### List Keyspaces

**GET** `/api/v1/schema/keyspaces`

Retrieve all keyspaces in the connected cluster.

**Request:** None

**Response:**
```json
{
  "keyspaces": [
    {
      "name": "system",
      "replication_strategy": "org.apache.cassandra.locator.LocalStrategy",
      "replication": {}
    },
    {
      "name": "app_data",
      "replication_strategy": "org.apache.cassandra.locator.NetworkTopologyStrategy",
      "replication": {
        "dc1": "3",
        "dc2": "3"
      }
    }
  ]
}
```

**Requires:** Authorization header

**Status Codes:**
- `200`: Success
- `401`: Unauthorized
- `500`: Server error

---

### List Tables

**GET** `/api/v1/schema/keyspaces/{keyspace}/tables`

Retrieve all tables in a specific keyspace.

**Path Parameters:**
- `keyspace` (string, required): Keyspace name

**Request:** None

**Response:**
```json
{
  "tables": [
    {
      "name": "users",
      "keyspace": "app_data",
      "estimated_rows": 1000000
    },
    {
      "name": "orders",
      "keyspace": "app_data",
      "estimated_rows": 5000000
    }
  ]
}
```

**Requires:** Authorization header

**Status Codes:**
- `200`: Success
- `401`: Unauthorized
- `404`: Keyspace not found
- `500`: Server error

---

### Get Table Schema

**GET** `/api/v1/schema/keyspaces/{keyspace}/tables/{table}`

Retrieve detailed schema information for a specific table.

**Path Parameters:**
- `keyspace` (string, required): Keyspace name
- `table` (string, required): Table name

**Request:** None

**Response:**
```json
{
  "schema": {
    "keyspace": "app_data",
    "table": "users",
    "columns": [
      {
        "name": "id",
        "type": "uuid",
        "is_partition_key": true,
        "is_clustering_key": false,
        "position": 0
      },
      {
        "name": "email",
        "type": "text",
        "is_partition_key": false,
        "is_clustering_key": false,
        "position": 1
      },
      {
        "name": "created_at",
        "type": "timestamp",
        "is_partition_key": false,
        "is_clustering_key": true,
        "position": 0
      }
    ],
    "partition_keys": ["id"],
    "clustering_keys": ["created_at"]
  }
}
```

**Requires:** Authorization header

**Status Codes:**
- `200`: Success
- `401`: Unauthorized
- `404`: Table not found
- `500`: Server error

---

## DataService

Provides data access and pagination for table rows.

### Query Rows

**POST** `/api/v1/data/query`

Query rows from a table with pagination.

**Request:**
```json
{
  "keyspace": "app_data",
  "table": "users",
  "page_size": 100
}
```

**Response:**
```json
{
  "rows": [
    {
      "cells": {
        "id": {
          "string_val": "550e8400-e29b-41d4-a716-446655440000",
          "is_null": false
        },
        "email": {
          "string_val": "user@example.com",
          "is_null": false
        },
        "created_at": {
          "string_val": "2024-01-15T10:30:00Z",
          "is_null": false
        }
      }
    }
  ],
  "cursor_id": "cursor_abc123",
  "has_more": true,
  "total_fetched": 100
}
```

**Requires:** Authorization header

**Status Codes:**
- `200`: Success
- `400`: Invalid request (missing keyspace/table)
- `401`: Unauthorized
- `404`: Table not found
- `500`: Server error

---

### Get Next Page

**POST** `/api/v1/data/next`

Fetch the next page of results using a cursor.

**Request:**
```json
{
  "cursor_id": "cursor_abc123"
}
```

**Response:**
```json
{
  "rows": [
    {
      "cells": {
        "id": {
          "string_val": "650e8400-e29b-41d4-a716-446655440000",
          "is_null": false
        },
        "email": {
          "string_val": "user2@example.com",
          "is_null": false
        }
      }
    }
  ],
  "cursor_id": "cursor_def456",
  "has_more": true
}
```

**Requires:** Authorization header

**Status Codes:**
- `200`: Success
- `400`: Invalid cursor ID
- `401`: Unauthorized
- `404`: Cursor not found or expired
- `500`: Server error

**Note:** Cursors expire after 30 minutes of inactivity.

---

### Filter Rows

**POST** `/api/v1/data/filter`

Query rows with a CQL WHERE clause filter.

**Request:**
```json
{
  "keyspace": "app_data",
  "table": "users",
  "where_clause": "email = 'user@example.com'",
  "page_size": 50
}
```

**Response:**
```json
{
  "rows": [
    {
      "cells": {
        "id": {
          "string_val": "550e8400-e29b-41d4-a716-446655440000",
          "is_null": false
        },
        "email": {
          "string_val": "user@example.com",
          "is_null": false
        }
      }
    }
  ],
  "cursor_id": "cursor_filter_xyz789",
  "has_more": false
}
```

**Requires:** Authorization header

**Status Codes:**
- `200`: Success
- `400`: Invalid WHERE clause syntax
- `401`: Unauthorized
- `404`: Table not found
- `500`: Server error

**Note:** WHERE clause must be valid CQL syntax without the `WHERE` keyword.

---

## Common Data Types

### CellValue

Represents a single cell value in a row. Uses oneof for type-safe value encoding.

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

**Cassandra Type Mappings:**
- `text`, `varchar`, `ascii`, `uuid`, `timeuuid`, `inet` → `string_val`
- `int`, `bigint`, `smallint`, `tinyint`, `counter`, `timestamp` → `int_val`
- `float`, `double`, `decimal` → `double_val`
- `boolean` → `bool_val`
- `blob`, `varint` → `bytes_val`
- `null` → `is_null = true`

---

### Column

Describes a table column.

```json
{
  "name": "user_id",
  "type": "uuid",
  "is_partition_key": true,
  "is_clustering_key": false,
  "position": 0
}
```

**Fields:**
- `name`: Column name
- `type`: CQL data type (e.g., `text`, `int`, `uuid`)
- `is_partition_key`: True if part of partition key
- `is_clustering_key`: True if part of clustering key
- `position`: Position in key (0-based)

---

## Error Handling

All errors follow a consistent structure:

```json
{
  "error": {
    "code": "INVALID_QUERY",
    "message": "Invalid WHERE clause syntax",
    "details": {
      "query": "invalid syntax here",
      "line": "1",
      "column": "10"
    }
  }
}
```

**Common Error Codes:**
- `UNAUTHENTICATED`: Missing or invalid authentication token
- `PERMISSION_DENIED`: Insufficient permissions for operation
- `NOT_FOUND`: Resource (keyspace/table/cursor) not found
- `INVALID_ARGUMENT`: Invalid request parameters
- `INVALID_QUERY`: Invalid CQL syntax in filter
- `INTERNAL`: Server internal error
- `UNAVAILABLE`: Database connection failed

---

## Rate Limits

No rate limits are currently enforced. Future versions may implement per-session limits.

---

## gRPC API

For gRPC clients, import the proto definitions from:

```
github.com/KashifKhn/kassie/api/proto/*.proto
```

**Example (Go):**
```go
import (
    kassiev1 "github.com/KashifKhn/kassie/api/gen/go"
    "google.golang.org/grpc"
)

conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := kassiev1.NewSessionServiceClient(conn)
resp, err := client.Login(ctx, &kassiev1.LoginRequest{
    Profile: "local",
})
```

---

## Client SDKs

Official client SDKs:
- **Go**: Built-in at `internal/client/client.go`
- **TypeScript**: Generated from proto (coming with Web UI)

Community SDKs:
- Python, Rust, Java - Coming soon

---

## OpenAPI Specification

An OpenAPI/Swagger specification can be generated from proto files:

```bash
protoc --openapiv2_out=docs/ \
  --openapiv2_opt=logtostderr=true \
  api/proto/*.proto
```

*(OpenAPI doc generation planned for future release)*

---

## Next Steps

- [CLI Commands](/reference/cli-commands) - Command-line interface
- [Configuration Schema](/reference/configuration-schema) - Config reference
- [Architecture](/architecture/) - System design
