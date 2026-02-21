# Advanced Filtering

Kassie supports CQL WHERE clauses for filtering table data. This guide covers common filter patterns and operators.

## Using the Filter Bar

### TUI

Press `/` to open the filter bar, type your WHERE clause, and press `Enter` to apply.

Press `Esc` to cancel, or clear the filter bar and press `Enter` to remove the filter.

### Web

The filter bar sits above the data grid. Type your WHERE clause and press `Enter`. Use the clear button (×) to remove the filter.

### API

```bash
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "keyspace": "app_data",
    "table": "users",
    "where_clause": "id = '\''550e8400-e29b-41d4-a716-446655440000'\''",
    "page_size": 100
  }' \
  http://localhost:8080/api/v1/data/filter | jq
```

## Filter by Partition Key

The most common and efficient filter — queries a single partition:

```cql
id = '550e8400-e29b-41d4-a716-446655440000'
```

::: tip
Filtering by partition key is always fast because Cassandra routes directly to the responsible node.
:::

## Filter by Clustering Key

Filter within a partition using clustering columns:

```cql
id = '550e8400-...' AND created_at > '2024-01-01'
```

```cql
id = '550e8400-...' AND created_at >= '2024-01-01' AND created_at < '2024-02-01'
```

::: warning
Clustering key filters require the partition key to be specified. Filtering by clustering columns only will fail.
:::

## Comparison Operators

| Operator | Example | Notes |
|----------|---------|-------|
| `=` | `status = 'active'` | Exact match |
| `<` | `age < 30` | Less than |
| `>` | `score > 100` | Greater than |
| `<=` | `created_at <= '2024-12-31'` | Less than or equal |
| `>=` | `priority >= 5` | Greater than or equal |

## Using the IN Operator

Query multiple partition key values:

```cql
id IN ('550e8400-...', '6ba7b810-...', '7c9e6679-...')
```

::: tip
`IN` on the partition key sends parallel queries to different nodes — very efficient for small lists.
:::

## Using CONTAINS

Filter collections (lists, sets, maps):

```cql
tags CONTAINS 'urgent'
```

```cql
properties CONTAINS KEY 'color'
```

::: warning
`CONTAINS` queries may require `ALLOW FILTERING` and can be slow on large datasets. Cassandra may reject these queries if it estimates a full table scan is needed.
:::

## Combining Filters

Combine multiple conditions with `AND`:

```cql
user_id = '550e8400-...' AND status = 'active' AND created_at > '2024-01-01'
```

::: warning
Cassandra has strict rules about which columns can be filtered together. Generally:
- Partition key columns must be specified with `=` or `IN`
- Clustering columns must follow the clustering order
- Non-key columns may require secondary indexes
:::

## Common Filter Examples

### Users table

```cql
-- Find a specific user
id = '550e8400-e29b-41d4-a716-446655440000'

-- Find users created in 2024
id = '550e8400-...' AND created_at >= '2024-01-01' AND created_at < '2025-01-01'
```

### Orders table

```cql
-- Find orders for a customer
customer_id = '6ba7b810-9dad-11d1-80b4-00c04fd430c8'

-- Find recent orders
customer_id = '6ba7b810-...' AND order_date > '2024-06-01'
```

### Logs table

```cql
-- Find errors for a service
service = 'payment-service' AND level = 'ERROR'

-- Find logs in a time range
service = 'api-gateway' AND timestamp >= '2024-06-15T00:00:00Z' AND timestamp < '2024-06-16T00:00:00Z'
```

## Filter Validation

Kassie validates filter syntax before sending to the database:

| Error | Cause | Fix |
|-------|-------|-----|
| Syntax error | Invalid CQL syntax | Check quotes and operators |
| Missing partition key | Non-key filter without PK | Add partition key condition |
| Invalid column name | Column doesn't exist | Check table schema |
| Type mismatch | Wrong value type for column | Match value type to column type |

Validation errors are displayed inline in the filter bar (both TUI and Web).

## Tips

1. **Always start with the partition key** — it's the most efficient filter
2. **Use the inspector** to check column types before writing filters
3. **Quote string values** — `name = 'Alice'` not `name = Alice`
4. **UUID values need quotes** — `id = '550e8400-...'`
5. **Timestamps use ISO 8601** — `created_at > '2024-01-01T00:00:00Z'`
6. **Clear the filter** to return to unfiltered browsing
