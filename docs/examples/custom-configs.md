# Custom Configurations

Examples of configuring Kassie for different environments and workflows.

## Config File Location

Kassie looks for configuration in this order:

1. `--config` flag path
2. `~/.config/kassie/config.json`
3. `./kassie.config.json` (current directory)

## Multiple Environment Profiles

Configure separate profiles for development, staging, and production:

```json
{
  "version": "1.0",
  "profiles": [
    {
      "name": "local",
      "hosts": ["127.0.0.1"],
      "port": 9042
    },
    {
      "name": "staging",
      "hosts": ["staging-db-1.internal", "staging-db-2.internal"],
      "port": 9042,
      "keyspace": "app_staging",
      "auth": {
        "username": "readonly",
        "password": "${STAGING_DB_PASSWORD}"
      }
    },
    {
      "name": "production",
      "hosts": ["prod-db-1.example.com", "prod-db-2.example.com", "prod-db-3.example.com"],
      "port": 9042,
      "keyspace": "app_production",
      "auth": {
        "username": "readonly",
        "password": "${PROD_DB_PASSWORD}"
      },
      "ssl": {
        "enabled": true,
        "cert_path": "/etc/kassie/client.crt",
        "key_path": "/etc/kassie/client.key",
        "ca_path": "/etc/kassie/ca.crt"
      }
    }
  ],
  "defaults": {
    "default_profile": "local",
    "page_size": 100,
    "timeout_ms": 5000
  }
}
```

### Usage

```bash
# Uses default profile (local)
kassie tui

# Specify profile
kassie tui --profile staging
kassie web --profile production
```

## SSL/TLS Configuration

### Basic SSL

```json
{
  "profiles": [{
    "name": "ssl-cluster",
    "hosts": ["db.example.com"],
    "port": 9142,
    "ssl": {
      "enabled": true
    }
  }]
}
```

### Mutual TLS (mTLS) with Client Certificates

```json
{
  "profiles": [{
    "name": "mtls-cluster",
    "hosts": ["db.example.com"],
    "port": 9142,
    "ssl": {
      "enabled": true,
      "cert_path": "${HOME}/.cassandra/client.crt",
      "key_path": "${HOME}/.cassandra/client.key",
      "ca_path": "${HOME}/.cassandra/ca.crt",
      "insecure_skip_verify": false
    }
  }]
}
```

### Skip Certificate Verification (Development Only)

::: warning
Only use this for local development with self-signed certificates.
:::

```json
{
  "ssl": {
    "enabled": true,
    "insecure_skip_verify": true
  }
}
```

## Custom Page Sizes

Adjust how many rows are fetched per page:

```json
{
  "defaults": {
    "page_size": 50
  }
}
```

| Page Size | Best For |
|-----------|----------|
| 50 | Quick browsing, slow connections |
| 100 | General use (default) |
| 500 | Reviewing large datasets |

## Timeout Adjustments

Configure query timeouts (in milliseconds):

```json
{
  "defaults": {
    "timeout_ms": 10000
  }
}
```

| Timeout | Best For |
|---------|----------|
| 3000 | Fast local clusters |
| 5000 | Default |
| 10000 | Cross-region clusters |
| 30000 | Very slow or loaded clusters |

## Theme Configuration

### TUI Themes

```json
{
  "clients": {
    "tui": {
      "theme": "dracula"
    }
  }
}
```

Available themes: `default`, `dracula`, `nord`, `gruvbox`

### Vim Mode

Enable Vim-style navigation:

```json
{
  "clients": {
    "tui": {
      "vim_mode": true
    }
  }
}
```

## Web Interface Options

### Custom Port

```json
{
  "clients": {
    "web": {
      "default_port": 9090
    }
  }
}
```

### Disable Auto-Open Browser

```json
{
  "clients": {
    "web": {
      "auto_open_browser": false
    }
  }
}
```

## Environment Variables

### Password Interpolation

Use `${VAR_NAME}` syntax to reference environment variables in passwords:

```json
{
  "auth": {
    "username": "readonly",
    "password": "${CASSANDRA_PASSWORD}"
  }
}
```

```bash
export CASSANDRA_PASSWORD="my-secret-password"
kassie tui --profile staging
```

::: tip
Environment variable interpolation keeps sensitive credentials out of config files. Use this for all passwords in shared or version-controlled configurations.
:::

### Path Interpolation

Environment variables also work in SSL certificate paths:

```json
{
  "ssl": {
    "cert_path": "${HOME}/.cassandra/client.crt",
    "key_path": "${HOME}/.cassandra/client.key"
  }
}
```

## Complete Example

A full configuration demonstrating all options:

```json
{
  "version": "1.0",
  "profiles": [
    {
      "name": "local",
      "hosts": ["127.0.0.1"],
      "port": 9042,
      "keyspace": "",
      "auth": {
        "username": "",
        "password": ""
      },
      "ssl": {
        "enabled": false
      }
    }
  ],
  "defaults": {
    "default_profile": "local",
    "page_size": 100,
    "timeout_ms": 5000
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

## Override Priority

Settings are resolved in this order (highest priority first):

1. **CLI flags** (`--profile`, `--port`)
2. **Client config** (`clients.tui.*`, `clients.web.*`)
3. **Profile config** (`profiles[n].*`)
4. **Defaults** (`defaults.*`)
5. **Built-in defaults**
