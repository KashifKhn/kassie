# Error Codes

Complete reference of error codes in Kassie.

## Error Code Format

Errors in Kassie follow this format:
```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message",
  "details": {}
}
```

## Authentication Errors

| Code | Description | Solution |
|------|-------------|----------|
| `AUTH_REQUIRED` | No authentication token provided | Login first |
| `AUTH_INVALID` | Token is invalid or malformed | Refresh token or login again |
| `AUTH_EXPIRED` | Token has expired | Use refresh token or login again |
| `AUTH_FORBIDDEN` | Insufficient permissions | Check user permissions |

## Connection Errors

| Code | Description | Solution |
|------|-------------|----------|
| `CONNECTION_FAILED` | Cannot connect to database | Check host, port, and database status |
| `CONNECTION_TIMEOUT` | Connection attempt timed out | Check network and increase timeout |
| `SSL_ERROR` | SSL/TLS connection failed | Verify SSL configuration |

## Configuration Errors

| Code | Description | Solution |
|------|-------------|----------|
| `PROFILE_NOT_FOUND` | Requested profile doesn't exist | Check profile name in config |
| `CONFIG_INVALID` | Configuration file is invalid | Validate JSON syntax |
| `CONFIG_NOT_FOUND` | Config file not found | Create config file |

## Query Errors

| Code | Description | Solution |
|------|-------------|----------|
| `QUERY_ERROR` | CQL query execution failed | Check query syntax |
| `QUERY_TIMEOUT` | Query exceeded timeout | Reduce data or increase timeout |
| `INVALID_FILTER` | WHERE clause syntax error | Fix filter syntax |
| `CURSOR_EXPIRED` | Pagination cursor expired | Refresh and start from first page |

## Internal Errors

| Code | Description | Solution |
|------|-------------|----------|
| `INTERNAL` | Unexpected server error | Check server logs, report if persistent |

See [Troubleshooting Guide](/guide/troubleshooting) for solutions to common errors.
