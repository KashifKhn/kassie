# Configuration Schema

See the [Configuration Guide](/guide/configuration) for complete documentation.

This page provides a technical schema reference.

## JSON Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["version", "profiles"],
  "properties": {
    "version": {
      "type": "string",
      "const": "1.0"
    },
    "profiles": {
      "type": "array",
      "items": {
        "$ref": "#/definitions/profile"
      }
    },
    "defaults": {
      "$ref": "#/definitions/defaults"
    },
    "clients": {
      "$ref": "#/definitions/clients"
    }
  }
}
```

## Field Reference

### Root Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | Yes | Configuration schema version (must be `"1.0"`) |
| `profiles` | array | Yes | List of database connection profiles |
| `defaults` | object | No | Default settings for connections and queries |
| `clients` | object | No | Client-specific settings (TUI and Web) |

### Profile

Database connection profile.

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `name` | string | Yes | Unique profile identifier | Must be unique |
| `hosts` | string[] | Yes | List of Cassandra/ScyllaDB host addresses | At least one host required |
| `port` | integer | Yes | CQL port number | Range: 1-65535 |
| `keyspace` | string | No | Default keyspace to connect to | - |
| `auth` | object | No | Authentication credentials | See `AuthConfig` |
| `ssl` | object | No | SSL/TLS configuration | See `SSLConfig` |

**Example**:
```json
{
  "name": "production",
  "hosts": ["db1.example.com", "db2.example.com"],
  "port": 9042,
  "keyspace": "myapp",
  "auth": {
    "username": "cassandra",
    "password": "${DB_PASSWORD}"
  }
}
```

### AuthConfig

Authentication credentials for the database.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `username` | string | Yes | Database username |
| `password` | string | Yes | Database password (supports environment variable interpolation) |

**Example**:
```json
{
  "username": "admin",
  "password": "${CASSANDRA_PASSWORD}"
}
```

### SSLConfig

SSL/TLS connection settings.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | boolean | Yes | - | Enable SSL/TLS connection |
| `cert_path` | string | No | - | Path to client certificate file (supports env interpolation) |
| `key_path` | string | No | - | Path to client private key file (supports env interpolation) |
| `ca_path` | string | No | - | Path to CA certificate file (supports env interpolation) |
| `insecure_skip_verify` | boolean | No | false | Skip certificate verification (insecure, use only for testing) |

**Example**:
```json
{
  "enabled": true,
  "cert_path": "${HOME}/.cassandra/client-cert.pem",
  "key_path": "${HOME}/.cassandra/client-key.pem",
  "ca_path": "${HOME}/.cassandra/ca-cert.pem",
  "insecure_skip_verify": false
}
```

### DefaultConfig

Default settings for database operations.

| Field | Type | Required | Default | Description | Validation |
|-------|------|----------|---------|-------------|------------|
| `default_profile` | string | No | First profile | Profile to use when none specified | Must reference existing profile |
| `page_size` | integer | No | 100 | Number of rows per page in query results | Range: 1-10000 |
| `timeout_ms` | integer | No | 10000 | Query timeout in milliseconds | Range: 100-300000 |

**Example**:
```json
{
  "default_profile": "local",
  "page_size": 100,
  "timeout_ms": 10000
}
```

### ClientConfig

Client-specific settings for TUI and Web interfaces.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tui` | object | No | Terminal UI settings (see `TUIConfig`) |
| `web` | object | No | Web UI settings (see `WebConfig`) |

### TUIConfig

Terminal UI configuration.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `theme` | string | No | `"default"` | Color theme for TUI |
| `vim_mode` | boolean | No | false | Enable Vim-style keybindings |

**Example**:
```json
{
  "theme": "default",
  "vim_mode": false
}
```

### WebConfig

Web UI configuration.

| Field | Type | Required | Default | Description | Validation |
|-------|------|----------|---------|-------------|------------|
| `auto_open_browser` | boolean | No | true | Automatically open browser when starting web UI |
| `default_port` | integer | No | 8080 | Default HTTP port for web UI | Range: 1-65535 |

**Example**:
```json
{
  "auto_open_browser": true,
  "default_port": 8080
}
```

## Environment Variable Interpolation

Configuration values support environment variable interpolation using the `${VAR_NAME}` syntax.

### Supported Fields

The following fields support environment variable interpolation:

- `auth.username`
- `auth.password`
- `ssl.cert_path`
- `ssl.key_path`
- `ssl.ca_path`

### Syntax

```json
{
  "auth": {
    "username": "${DB_USER}",
    "password": "${DB_PASSWORD}"
  },
  "ssl": {
    "cert_path": "${HOME}/.cassandra/cert.pem"
  }
}
```

### Rules

- Pattern: `${VARIABLE_NAME}`
- Variable names must match regex: `[A-Z_][A-Z0-9_]*`
- Variables are resolved at runtime when config is loaded
- Nested interpolation supported (up to 10 levels)
- Circular references detected and rejected
- Missing variables cause config load failure

### Examples

**Using credentials from environment**:
```bash
export CASSANDRA_USER="admin"
export CASSANDRA_PASS="secret123"
```

```json
{
  "profiles": [{
    "name": "production",
    "hosts": ["prod.example.com"],
    "port": 9042,
    "auth": {
      "username": "${CASSANDRA_USER}",
      "password": "${CASSANDRA_PASS}"
    }
  }]
}
```

**Using path expansion**:
```json
{
  "profiles": [{
    "ssl": {
      "enabled": true,
      "cert_path": "${HOME}/.cassandra/client.crt",
      "key_path": "${HOME}/.cassandra/client.key"
    }
  }]
}
```

## Validation Rules

Configuration is validated when loaded. Validation failures prevent startup.

### Profile Validation

- **Name**: Must be non-empty string
- **Hosts**: At least one host required
- **Port**: Must be in range 1-65535
- **Profile names**: Must be unique across all profiles

### Defaults Validation

- **PageSize**: Must be in range 1-10000
- **TimeoutMs**: Must be in range 100-300000 (0.1s - 5min)

### Clients Validation

- **Web.DefaultPort**: Must be in range 1-65535

### Error Messages

| Validation Error | Cause |
|------------------|-------|
| `profile not found` | Referenced profile does not exist |
| `invalid port number` | Port outside valid range (1-65535) |
| `no hosts specified` | Profile has empty hosts array |
| `invalid configuration` | Profile missing required name field |
| `duplicate profile name` | Two profiles have the same name |
| `no profiles defined` | Config has empty profiles array |
| `invalid page size` | PageSize outside range 1-10000 |
| `invalid timeout` | TimeoutMs outside range 100-300000 |

## Complete Example

```json
{
  "version": "1.0",
  "profiles": [
    {
      "name": "local",
      "hosts": ["127.0.0.1"],
      "port": 9042,
      "keyspace": "system"
    },
    {
      "name": "production",
      "hosts": ["db1.prod.example.com", "db2.prod.example.com"],
      "port": 9042,
      "keyspace": "myapp",
      "auth": {
        "username": "${PROD_DB_USER}",
        "password": "${PROD_DB_PASSWORD}"
      },
      "ssl": {
        "enabled": true,
        "ca_path": "${HOME}/.cassandra/ca.pem",
        "insecure_skip_verify": false
      }
    }
  ],
  "defaults": {
    "default_profile": "local",
    "page_size": 100,
    "timeout_ms": 10000
  },
  "clients": {
    "tui": {
      "theme": "default",
      "vim_mode": false
    },
    "web": {
      "auto_open_browser": true,
      "default_port": 8080
    }
  }
}
```

See the [Configuration Guide](/guide/configuration) for complete documentation.
