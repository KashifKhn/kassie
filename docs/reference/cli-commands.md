# CLI Commands

Complete reference for all Kassie command-line interfaces.

## Global Options

These flags work with all commands:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--config` | - | string | Path to config file (default: `~/.config/kassie/config.json`) |
| `--profile` | - | string | Database profile to use |
| `--log-level` | - | string | Log level: `debug`, `info`, `warn`, `error` (default: `info`) |
| `--version` | `-v` | boolean | Print version information and exit |
| `--help` | `-h` | boolean | Show help message |

**Examples**:
```bash
# Use custom config
kassie tui --config /path/to/config.json

# Set log level
kassie tui --log-level debug

# Use specific profile
kassie tui --profile production
```

## Commands

### `kassie tui`

Launch the Terminal User Interface.

**Usage**:
```bash
kassie tui [options]
```

**Options**:

| Flag | Type | Description |
|------|------|-------------|
| `--profile` | string | Profile to connect to |
| `--server` | string | Connect to remote Kassie server (format: `host:port`) |

**Examples**:
```bash
# Launch TUI with default settings
kassie tui

# Connect to remote server
kassie tui --server remote.example.com:50051

# Use specific profile
kassie tui --profile production

# Debug mode
kassie tui --log-level debug
```

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Configuration error
- `3`: Connection error

---

### `kassie web`

Launch the Web User Interface.

**Usage**:
```bash
kassie web [options]
```

**Options**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | integer | 8080 | HTTP port for web server |
| `--no-browser` | boolean | false | Don't auto-open browser |
| `--profile` | string | - | Default profile to connect |

**Examples**:
```bash
# Launch web UI on default port (8080)
kassie web

# Use custom port
kassie web --port 3000

# Don't open browser automatically
kassie web --no-browser
```

**Access**:
- Default URL: `http://localhost:8080`
- Custom port: `http://localhost:<port>`

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Configuration error
- `4`: Port already in use

---

### `kassie server`

Run Kassie in standalone server mode.

**Usage**:
```bash
kassie server [options]
```

**Options**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--grpc-port` | integer | 50051 | gRPC server port (fixed) |
| `--http-port` | integer | 8080 | HTTP gateway port (fixed) |
| `--host` | string | `0.0.0.0` | Bind address |

**Port Behavior**:
- Server mode uses **fixed ports** specified by flags
- Both gRPC and HTTP servers bind to the same host address
- Ports must not be in use or server will fail to start

::: info Embedded vs Standalone Ports
- **`kassie server`**: Uses fixed ports (50051, 8080) - suitable for production
- **`kassie tui` / `kassie web`**: Uses dynamic ports (random free ports) for embedded server
:::

**Examples**:
```bash
# Run server with default ports
kassie server

# Custom ports
kassie server --grpc-port 9090 --http-port 8888

# Bind to specific interface
kassie server --host 192.168.1.100

# Production mode
kassie server --log-level warn
```

**Endpoints**:
- gRPC: `<host>:<grpc-port>`
- HTTP: `http://<host>:<http-port>`
- Health check: `http://<host>:<http-port>/health`

**Signals**:
- `SIGINT` / `SIGTERM`: Graceful shutdown
- `SIGKILL`: Force shutdown (not recommended)

---

### `kassie upgrade`

Upgrade Kassie to the latest version or a specific version.

**Usage**:
```bash
kassie upgrade [options]
```

**Options**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--check` | `-c` | boolean | false | Only check for updates without installing |
| `--force` | `-f` | boolean | false | Force upgrade even if already on latest |
| `--version` | `-v` | string | - | Upgrade to a specific version |
| `--json` | - | boolean | false | Output result as JSON |

**Examples**:
```bash
# Check and upgrade to latest version
kassie upgrade

# Only check for updates
kassie upgrade --check

# Check with JSON output
kassie upgrade --check --json

# Force reinstall current version
kassie upgrade --force

# Upgrade to specific version
kassie upgrade --version v0.1.1

# Downgrade to older version
kassie upgrade --version v0.1.0
```

**How It Works**:

The upgrade command:
1. Checks GitHub releases for the target version
2. Downloads the appropriate binary for your platform
3. Verifies the download using SHA256 checksums
4. Creates a backup of your current binary
5. Installs the new version to all detected installation locations
6. Verifies the new installation works
7. Automatically rolls back if anything fails

**JSON Output Format**:
```json
{
  "current_version": "v0.1.0",
  "latest_version": "v0.1.1",
  "update_available": true,
  "upgraded": false,
  "platform": {
    "os": "linux",
    "arch": "amd64"
  },
  "installed_paths": [
    "/usr/local/bin/kassie"
  ]
}
```

**Safety Features**:
- Automatic backup before upgrade
- Checksum verification of downloads
- Installation verification before committing
- Automatic rollback on failure
- Multi-location installation support

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Download failed
- `3`: Verification failed
- `4`: Installation failed

---

### `kassie version`

Print version information.

**Usage**:
```bash
kassie version
```

**Output**:

<VersionInfo />

**Programmatic access**:
```bash
# Get just version number
kassie version | head -1 | awk '{print $2}'
```

---

### `kassie help`

Show help information.

**Usage**:
```bash
kassie help [command]
```

**Examples**:
```bash
# General help
kassie help

# Help for specific command
kassie help tui
kassie help web
kassie help server
```

## Environment Variables

Kassie recognizes these environment variables:

| Variable | Description |
|----------|-------------|
| `KASSIE_CONFIG` | Config file path (overrides `--config`) |
| `KASSIE_PROFILE` | Default profile (overrides `--profile`) |
| `KASSIE_LOG_LEVEL` | Log level (overrides `--log-level`) |
| `KASSIE_JWT_SECRET` | JWT secret for authentication (see below) |

### JWT Secret Usage

The `KASSIE_JWT_SECRET` environment variable is used differently depending on the deployment mode:

**Standalone Server Mode (`kassie server`)**:
```bash
export KASSIE_JWT_SECRET="your-production-secret-here"
kassie server
```
- **Purpose:** Secures the standalone server
- **Requirement:** Set in production environments
- **Default:** `"change-this-secret-in-production"` (warning logged)
- **Security:** Use a strong, random secret (min 32 characters)

**Embedded Mode (`kassie tui` / `kassie web`)**:
```bash
# Optional - defaults are auto-generated per session
export KASSIE_JWT_SECRET="my-local-secret"
kassie tui
```
- **Purpose:** Internal authentication between client and embedded server
- **Requirement:** Optional (auto-generated if not set)
- **Default:** `"tui-mode-secret"` for TUI, `"web-mode-secret"` for web
- **Security:** Less critical since server runs locally in same process

::: tip Security Recommendation
For `kassie server` in production, always set a strong `KASSIE_JWT_SECRET`:
```bash
# Generate a secure random secret
export KASSIE_JWT_SECRET=$(openssl rand -base64 32)
```
:::

**Example**:
```bash
export KASSIE_CONFIG=~/.kassie.json
export KASSIE_PROFILE=production
export KASSIE_LOG_LEVEL=debug
kassie tui
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Connection error |
| 4 | Port/resource unavailable |
| 5 | Permission denied |

## Common Patterns

### Quick connection
```bash
kassie tui --profile local
```

### Development with debug logging
```bash
kassie tui --log-level debug
```

### Production server
```bash
kassie server \
  --host 0.0.0.0 \
  --grpc-port 50051 \
  --http-port 8080 \
  --log-level info
```

### Team web access
```bash
kassie web --port 8080 --no-browser
```

### Remote TUI connection
```bash
kassie tui --server prod-kassie.example.com:50051
```

## Scripting

Kassie can be used in scripts:

```bash
#!/bin/bash

# Check version
VERSION=$(kassie version | head -1 | awk '{print $2}')
echo "Using Kassie $VERSION"

# Start server in background
kassie server --log-level warn > kassie.log 2>&1 &
PID=$!

# Wait for server to start
sleep 2

# Do work...

# Stop server
kill $PID
```

## Next Steps

- [Configuration Schema](/reference/configuration-schema) - Detailed config reference
- [Keyboard Shortcuts](/reference/keyboard-shortcuts) - All shortcuts
- [API Reference](/reference/api) - Programmatic access
