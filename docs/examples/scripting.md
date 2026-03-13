# Scripting & Automation

Use Kassie's REST API for scripting, health checks, and CI/CD integration.

## Authentication

All API requests require a JWT token. First, log in to get tokens:

```bash
# Login and extract access token
TOKEN=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{"profile": "local"}' \
  http://localhost:8080/api/v1/session/login | jq -r '.access_token')

echo "Token: $TOKEN"
```

Use the token in subsequent requests:

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/schema/keyspaces | jq
```

## Health Check Script

Verify Kassie can connect to a database:

```bash
#!/bin/bash
# health-check.sh — Check Kassie connectivity

HOST="${KASSIE_HOST:-localhost}"
PORT="${KASSIE_PORT:-8080}"
PROFILE="${KASSIE_PROFILE:-local}"

# Login
TOKEN=$(curl -sf -X POST \
  -H "Content-Type: application/json" \
  -d "{\"profile\": \"$PROFILE\"}" \
  "http://$HOST:$PORT/api/v1/session/login" | jq -r '.access_token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "FAIL: Cannot connect to Kassie or login failed"
  exit 1
fi

# Check keyspaces
KEYSPACES=$(curl -sf \
  -H "Authorization: Bearer $TOKEN" \
  "http://$HOST:$PORT/api/v1/schema/keyspaces" | jq '.keyspaces | length')

if [ "$KEYSPACES" -gt 0 ]; then
  echo "OK: Connected, found $KEYSPACES keyspaces"
  exit 0
else
  echo "FAIL: Connected but no keyspaces found"
  exit 1
fi
```

Usage:
```bash
chmod +x health-check.sh
./health-check.sh
```

## Data Validation Script

Verify expected data exists in a table:

```bash
#!/bin/bash
# validate-data.sh — Check that expected records exist

HOST="${1:-localhost:8080}"
PROFILE="${2:-local}"

# Login
TOKEN=$(curl -sf -X POST \
  -H "Content-Type: application/json" \
  -d "{\"profile\": \"$PROFILE\"}" \
  "http://$HOST/api/v1/session/login" | jq -r '.access_token')

AUTH="Authorization: Bearer $TOKEN"

# Check users table has data
RESPONSE=$(curl -sf -X POST \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"keyspace": "app_data", "table": "users", "page_size": 1}' \
  "http://$HOST/api/v1/data/query")

ROW_COUNT=$(echo "$RESPONSE" | jq '.rows | length')
HAS_MORE=$(echo "$RESPONSE" | jq '.has_more')

echo "Users table: $ROW_COUNT rows returned, has_more=$HAS_MORE"

if [ "$ROW_COUNT" -eq 0 ]; then
  echo "WARNING: Users table is empty!"
  exit 1
fi

# Validate specific record exists
FILTER_RESPONSE=$(curl -sf -X POST \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{
    "keyspace": "app_data",
    "table": "users",
    "where_clause": "id = '\''550e8400-e29b-41d4-a716-446655440000'\''",
    "page_size": 1
  }' \
  "http://$HOST/api/v1/data/filter")

FOUND=$(echo "$FILTER_RESPONSE" | jq '.rows | length')

if [ "$FOUND" -gt 0 ]; then
  echo "OK: Expected user record found"
else
  echo "FAIL: Expected user record not found"
  exit 1
fi

# Logout
curl -sf -X POST -H "$AUTH" "http://$HOST/api/v1/session/logout" > /dev/null

echo "Validation complete"
```

## Schema Inventory Script

List all tables across all keyspaces:

```bash
#!/bin/bash
# schema-inventory.sh — List all keyspaces and tables

HOST="${1:-localhost:8080}"
PROFILE="${2:-local}"

TOKEN=$(curl -sf -X POST \
  -H "Content-Type: application/json" \
  -d "{\"profile\": \"$PROFILE\"}" \
  "http://$HOST/api/v1/session/login" | jq -r '.access_token')

AUTH="Authorization: Bearer $TOKEN"

# Get all keyspaces
KEYSPACES=$(curl -sf -H "$AUTH" \
  "http://$HOST/api/v1/schema/keyspaces" | jq -r '.keyspaces[].name')

for KS in $KEYSPACES; do
  # Skip system keyspaces
  case "$KS" in
    system|system_schema|system_auth|system_distributed|system_traces)
      continue ;;
  esac

  echo "=== $KS ==="
  curl -sf -H "$AUTH" \
    "http://$HOST/api/v1/schema/keyspaces/$KS/tables" | \
    jq -r '.tables[] | "  \(.name) (~\(.estimated_rows) rows)"'
  echo
done

# Cleanup
curl -sf -X POST -H "$AUTH" "http://$HOST/api/v1/session/logout" > /dev/null
```

Output:
```
=== app_data ===
  users (~50000 rows)
  orders (~120000 rows)
  logs (~5000000 rows)
```

## CI/CD Integration

### GitHub Actions — Smoke Test

```yaml
# .github/workflows/smoke-test.yml
name: Database Smoke Test

on:
  workflow_dispatch:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours

jobs:
  smoke-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Start ScyllaDB
        run: docker-compose up -d

      - name: Wait for ScyllaDB
        run: sleep 30

      - name: Install Kassie
        run: go install github.com/KashifKhn/kassie@latest

      - name: Start Kassie server
        run: |
          kassie server --http-port 8080 &
          sleep 2

      - name: Run health check
        run: ./scripts/health-check.sh
```

### Docker — Run Against Remote Cluster

```bash
# Start Kassie server pointing to remote cluster
docker run -d \
  -p 8080:8080 \
  -v /path/to/config.json:/root/.config/kassie/config.json \
  ghcr.io/kashifkhn/kassie server --http-port 8080

# Run validation script against it
./validate-data.sh localhost:8080 production
```

## API Quick Reference

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/session/login` | POST | No | Login with profile |
| `/api/v1/session/refresh` | POST | No | Refresh access token |
| `/api/v1/session/logout` | POST | Yes | End session |
| `/api/v1/profiles` | GET | No | List profiles |
| `/api/v1/schema/keyspaces` | GET | Yes | List keyspaces |
| `/api/v1/schema/keyspaces/{ks}/tables` | GET | Yes | List tables |
| `/api/v1/schema/keyspaces/{ks}/tables/{tbl}` | GET | Yes | Get schema |
| `/api/v1/data/query` | POST | Yes | Query rows |
| `/api/v1/data/next` | POST | Yes | Next page |
| `/api/v1/data/filter` | POST | Yes | Filter rows |

## Tips

1. **Always logout** after scripts to clean up server sessions
2. **Use `jq`** for JSON parsing in shell scripts
3. **Set timeouts** on curl with `-m 10` to avoid hanging scripts
4. **Store tokens** in variables, not files (avoid token leakage)
5. **Check `has_more`** to paginate through full result sets
