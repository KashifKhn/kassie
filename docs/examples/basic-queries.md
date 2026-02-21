# Basic Queries

Learn the fundamentals of browsing data with Kassie using both TUI and Web interfaces.

## Viewing Keyspaces

### TUI

```bash
kassie tui --profile local
```

After connecting, the sidebar shows all keyspaces. Use `j`/`k` to navigate and `Enter` to expand.

### Web

```bash
kassie web --profile local
```

The sidebar displays keyspaces as an accordion tree. Click a keyspace to expand it.

### API

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/schema/keyspaces | jq
```

Response:
```json
{
  "keyspaces": [
    { "name": "system", "replication_strategy": "LocalStrategy", "replication": {} },
    { "name": "system_schema", "replication_strategy": "LocalStrategy", "replication": {} },
    { "name": "app_data", "replication_strategy": "SimpleStrategy", "replication": { "replication_factor": "3" } }
  ]
}
```

## Listing Tables

### TUI

Expand a keyspace in the sidebar with `Enter`. Tables appear nested under the keyspace.

### Web

Click a keyspace name to toggle its table list.

### API

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/schema/keyspaces/app_data/tables | jq
```

Response:
```json
{
  "tables": [
    { "name": "users", "keyspace": "app_data", "estimated_rows": 50000 },
    { "name": "orders", "keyspace": "app_data", "estimated_rows": 120000 },
    { "name": "logs", "keyspace": "app_data", "estimated_rows": 5000000 }
  ]
}
```

## Viewing Table Schema

### TUI

Select a table in the sidebar. The data grid header shows column names. Press `Enter` on any row to see column types in the inspector.

### Web

Select a table in the sidebar. The data grid displays column headers with type information. Click any row to open the inspector panel.

### API

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/schema/keyspaces/app_data/tables/users | jq
```

Response:
```json
{
  "schema": {
    "keyspace": "app_data",
    "table": "users",
    "columns": [
      { "name": "id", "type": "uuid", "is_partition_key": true, "position": 0 },
      { "name": "email", "type": "text", "is_partition_key": false, "position": 1 },
      { "name": "name", "type": "text", "is_partition_key": false, "position": 2 },
      { "name": "created_at", "type": "timestamp", "is_clustering_key": true, "position": 0 }
    ],
    "partition_keys": ["id"],
    "clustering_keys": ["created_at"]
  }
}
```

## Fetching Rows

### TUI

Select a table in the sidebar — the first page of rows loads automatically in the data grid.

### Web

Click a table name — rows appear in the center data grid panel.

### API

```bash
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"keyspace": "app_data", "table": "users", "page_size": 10}' \
  http://localhost:8080/api/v1/data/query | jq
```

Response:
```json
{
  "rows": [
    {
      "cells": {
        "id": { "stringVal": "550e8400-e29b-41d4-a716-446655440000" },
        "email": { "stringVal": "alice@example.com" },
        "name": { "stringVal": "Alice" },
        "created_at": { "stringVal": "2024-06-15T10:30:00Z" }
      }
    }
  ],
  "cursor_id": "cur-abc-123",
  "has_more": true,
  "total_fetched": 10
}
```

## Navigating Pages

### TUI

- Press `n` — Next page
- Press `p` — Previous page (returns to cached data)
- Status bar shows page info and row count

### Web

Use the pagination controls at the bottom of the data grid. The "Next" button fetches the next page via cursor.

### API

```bash
# Next page using cursor
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"cursor_id": "cur-abc-123"}' \
  http://localhost:8080/api/v1/data/next | jq
```

## Inspecting a Row

### TUI

Navigate to a row in the data grid and press `Enter`. The inspector panel shows:

- All column values in a formatted JSON tree
- Partition keys marked with `[PK]`
- Clustering keys marked with `[CK]`
- Column types displayed alongside values

Press `Esc` to close the inspector.

### Web

Click any row in the data grid. The right inspector panel displays:

- Collapsible JSON tree view
- Partition and clustering key badges
- Column types
- Copy to clipboard support

### Keyboard Shortcuts Summary

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate up/down |
| `Enter` | Select/expand |
| `n` | Next page |
| `p` | Previous page |
| `/` | Open filter bar |
| `Tab` | Switch panes |
| `Esc` | Close/cancel |
| `q` | Quit |
