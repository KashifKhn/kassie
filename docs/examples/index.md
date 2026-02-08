# Examples

Practical examples of using Kassie.

## Quick Examples

### Basic Queries
Learn how to browse keyspaces, tables, and rows.

```bash
kassie tui --profile local
# Navigate to system_schema keyspace
# Select tables table
# View all tables in your cluster
```

### Filtering
Apply WHERE clauses to find specific data.

```cql
# Filter by partition key
id = '550e8400-e29b-41d4-a716-446655440000'

# Range query
created_at > '2024-01-01' AND created_at < '2024-02-01'
```

### Custom Configuration
Configure profiles for different environments.

```json
{
  "profiles": [
    {
      "name": "local",
      "hosts": ["localhost"],
      "port": 9042
    },
    {
      "name": "production",
      "hosts": ["prod-1.example.com"],
      "port": 9042,
      "ssl": { "enabled": true }
    }
  ]
}
```

## Example Guides

- [Basic Queries](/examples/basic-queries) - Simple query examples
- [Advanced Filtering](/examples/advanced-filtering) - Complex filters
- [Custom Configurations](/examples/custom-configs) - Configuration examples
- [Scripting](/examples/scripting) - Automation examples

## More Examples

See the [Guide](/guide/) section for step-by-step tutorials.
