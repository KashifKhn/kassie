# Configuration

Kassie uses a JSON configuration file to manage database profiles and settings. This guide covers all configuration options.

## Configuration File Location

Kassie looks for configuration in the following order:

1. Path specified via `--config` flag
2. `~/.config/kassie/config.json` (primary location)
3. Built-in defaults (if no config file found)

::: info Note
Unlike earlier documentation, Kassie does **not** check for `./kassie.config.json` in the current directory. If you need a project-specific config, use the `--config` flag:
```bash
kassie tui --config ./project-config.json
```
:::

## Basic Configuration

Create `~/.config/kassie/config.json`:

```json
{
  "version": "1.0",
  "profiles": [
    {
      "name": "local",
      "hosts": ["127.0.0.1"],
      "port": 9042,
      "keyspace": "system"
    }
  ]
}
```

## Complete Configuration Example

Here's a full configuration with all available options:

```json
{
  "version": "1.0",
  "profiles": [
    {
      "name": "local",
      "hosts": ["127.0.0.1"],
      "port": 9042,
      "keyspace": "system",
      "auth": {
        "username": "cassandra",
        "password": "cassandra"
      },
      "ssl": {
        "enabled": false,
        "cert_path": "",
        "key_path": "",
        "ca_path": "",
        "insecure_skip_verify": false
      }
    },
    {
      "name": "production",
      "hosts": [
        "10.0.1.1",
        "10.0.1.2",
        "10.0.1.3"
      ],
      "port": 9042,
      "keyspace": "app_data",
      "auth": {
        "username": "admin",
        "password": "${PROD_CASSANDRA_PASSWORD}"
      },
      "ssl": {
        "enabled": true,
        "cert_path": "/path/to/client.crt",
        "key_path": "/path/to/client.key",
        "ca_path": "/path/to/ca.crt",
        "insecure_skip_verify": true
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

## Configuration Reference

### Root Level

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | Yes | Config version (currently "1.0") |
| `profiles` | array | Yes | List of database profiles |
| `defaults` | object | No | Default settings |
| `clients` | object | No | Client-specific overrides |

### Profile Configuration

Each profile in the `profiles` array can have:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Unique profile name |
| `hosts` | array | Yes | - | List of Cassandra/Scylla hosts |
| `port` | integer | No | 9042 | CQL port |
| `keyspace` | string | No | "system" | Default keyspace |
| `auth` | object | No | - | Authentication credentials |
| `ssl` | object | No | - | SSL/TLS configuration |

### Authentication

```json
{
  "auth": {
    "username": "cassandra",
    "password": "cassandra"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `username` | string | No | Database username |
| `password` | string | No | Database password (supports env vars) |

### SSL/TLS Configuration

```json
{
  "ssl": {
    "enabled": true,
    "cert_path": "/path/to/client.crt",
    "key_path": "/path/to/client.key",
    "ca_path": "/path/to/ca.crt",
    "insecure_skip_verify": false
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `enabled` | boolean | No | Enable SSL/TLS |
| `cert_path` | string | No | Client certificate path |
| `key_path` | string | No | Client key path |
| `ca_path` | string | No | CA certificate path |
| `insecure_skip_verify` | boolean | No | Skip certificate chain verification (insecure) |

::: warning
When `ssl.enabled` is `true`, you must provide valid certificate paths.
:::

### Defaults

```json
{
  "defaults": {
    "default_profile": "local",
    "page_size": 100,
    "timeout_ms": 10000
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `default_profile` | string | Profile to use if none specified |
| `page_size` | integer | Number of rows per page |
| `timeout_ms` | integer | Default query timeout |

### TUI Client Configuration

```json
{
  "clients": {
    "tui": {
      "theme": "default",
      "vim_mode": false
    }
  }
}
```

| Field | Type | Options | Description |
|-------|------|---------|-------------|
| `theme` | string | `default` | Color scheme (additional themes coming soon) |
| `vim_mode` | boolean | - | Enable Vim-style navigation |

::: info Theme Development
Currently only the `default` theme is fully implemented. Additional themes (`dracula`, `nord`, `gruvbox`) are planned for future releases.
:::

### Web Client Configuration

```json
{
  "clients": {
    "web": {
      "auto_open_browser": true,
      "default_port": 8080
    }
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `auto_open_browser` | boolean | Open browser automatically on launch |
| `default_port` | integer | Default HTTP port for web UI |

## Environment Variables

### Password Interpolation

Use environment variables for sensitive data:

```json
{
  "auth": {
    "username": "admin",
    "password": "${CASSANDRA_PASSWORD}"
  }
}
```

Set the environment variable:

```bash
export CASSANDRA_PASSWORD="secret123"
kassie tui --profile production
```

### Supported Variables

You can use environment variables in:
- `auth.password`
- `ssl.cert_path`
- `ssl.key_path`
- `ssl.ca_path`

## CLI Overrides

CLI flags override configuration file settings:

```bash
# Override profile
kassie tui --profile production

# Override config file location
kassie tui --config /path/to/custom-config.json

# Override log level
kassie tui --log-level debug
```

## Multiple Profiles

Create profiles for different environments:

```json
{
  "profiles": [
    {
      "name": "local",
      "hosts": ["localhost"],
      "port": 9042
    },
    {
      "name": "staging",
      "hosts": ["staging-db.example.com"],
      "port": 9042,
      "auth": {
        "username": "staging_user",
        "password": "${STAGING_PASSWORD}"
      }
    },
    {
      "name": "production",
      "hosts": [
        "prod-1.example.com",
        "prod-2.example.com",
        "prod-3.example.com"
      ],
      "port": 9042,
      "auth": {
        "username": "prod_user",
        "password": "${PROD_PASSWORD}"
      },
      "ssl": {
        "enabled": true,
        "ca_path": "/etc/ssl/certs/ca.crt"
      }
    }
  ],
  "defaults": {
    "default_profile": "local"
  }
}
```

Switch between profiles:

```bash
kassie tui --profile staging
kassie tui --profile production
```

## Advanced Configuration

### Connection Pooling

Kassie manages connection pools automatically. Each profile gets its own pool with:
- Default pool size: 5 connections
- Automatic reconnection on failures
- Health checks every 30 seconds

### Timeout Configuration

Configure timeouts globally in defaults:

```json
{
  "defaults": {
    "timeout_ms": 10000
  }
}
```

### Page Size

Control pagination with `page_size`:

```json
{
  "defaults": {
    "page_size": 50
  }
}
```

- Smaller values: Faster initial load, more frequent paging
- Larger values: Fewer pages, more memory usage
- Recommended: 50-200

## Validation

Kassie validates your configuration on startup. Common errors:

### Missing Required Fields

```
Error: profile 'local' missing required field 'hosts'
```

Fix: Add the required field to your profile.

### Invalid JSON

```
Error: invalid config: unexpected token at line 5
```

Fix: Validate your JSON syntax (use a JSON validator).

### Invalid Profile Name

```
Error: profile name must be alphanumeric and underscore only
```

Fix: Use only letters, numbers, and underscores in profile names.

## Configuration Examples

See [Custom Configurations](/examples/custom-configs) for more examples:
- Multi-datacenter setup
- SSL/TLS with client certificates
- Authentication with LDAP
- Custom timeouts and page sizes

## Next Steps

- [TUI Usage](/guide/tui-usage) - Learn how to use the terminal interface
- [Web Usage](/guide/web-usage) - Explore the web interface
- [Troubleshooting](/guide/troubleshooting) - Fix common configuration issues
