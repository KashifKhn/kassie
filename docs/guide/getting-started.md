# Getting Started

This guide will help you get Kassie up and running in under 5 minutes.

## Prerequisites

All you need is:
- Access to a Cassandra or ScyllaDB cluster
- The connection details (host, port, credentials)

That's it! Kassie has no other dependencies.

## Installation

Choose your preferred installation method:

::: code-group

```bash [Curl (Recommended)]
curl -fsSL https://raw.githubusercontent.com/KashifKhn/kassie/main/install.sh | sh
```

```bash [Go Install]
go install github.com/kashifKhn/kassie@latest
```

```bash [Homebrew (Coming Soon)]
# Homebrew tap planned for future release
brew tap KashifKhn/kassie
brew install kassie
```

:::

Verify the installation:

```bash
kassie version
```

::: warning Web UI Under Development
The web interface (`kassie web`) is currently in active development. For production use, we recommend the TUI (Terminal UI) interface which is fully functional and stable.
:::

## First Run

### Step 1: Create a Configuration File

Create a config file at `~/.config/kassie/config.json`:

```bash
mkdir -p ~/.config/kassie
```

Add your database connection details:

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
  ],
  "defaults": {
    "default_profile": "local",
    "page_size": 100,
    "timeout_ms": 10000
  }
}
```

::: tip
If you don't create a config file, Kassie will use sensible defaults and try to connect to `127.0.0.1:9042`.
:::

### Step 2: Launch Kassie

Start the TUI:

```bash
kassie tui
```

::: tip Recommended
The TUI provides the most stable and feature-complete experience. The web interface is under active development.
:::

## First Steps in TUI

Once Kassie starts, you'll see the connection view:

1. **Select a Profile**: Use `j/k` or arrow keys to navigate
2. **Connect**: Press `Enter` to connect to the selected profile
3. **Explore**: You'll see a sidebar with all keyspaces

### Navigate the Interface

The TUI has three main panels:

```
┌─────────────────┬──────────────────────────────────────┐
│                 │                                      │
│    Sidebar      │        Data Grid                    │
│                 │                                      │
│  Keyspaces      │    Table Rows                       │
│  └─ Tables      │                                      │
│                 │                                      │
├─────────────────┼──────────────────────────────────────┤
│                 │                                      │
│                 │        Inspector                     │
│                 │    (Selected Row Details)            │
│                 │                                      │
└─────────────────┴──────────────────────────────────────┘
```

### Browse Data

1. **Navigate Keyspaces**: Use `j/k` to move up/down in the sidebar
2. **Expand Keyspace**: Press `l` or `Enter` to expand and see tables
3. **Select Table**: Navigate to a table and press `Enter`
4. **View Data**: The data grid shows the table's rows
5. **Page Through Data**: Press `n` for next page, `p` for previous
6. **Inspect Row**: Select a row and press `Enter` to see details

### Apply Filters

Press `/` to open the filter bar and type a WHERE clause:

```cql
id = '550e8400-e29b-41d4-a716-446655440000'
```

Press `Enter` to apply the filter.

## Quick Tutorial: Exploring System Tables

Let's explore Cassandra's system tables:

1. Launch Kassie: `kassie tui`
2. Connect to your local cluster
3. Navigate to the `system_schema` keyspace
4. Expand it with `l` or `Enter`
5. Select the `tables` table
6. Press `Enter` to view all tables in your cluster
7. Use `n` to page through results
8. Select a row and press `Enter` to inspect its details

## Key Shortcuts

| Key | Action |
|-----|--------|
| `j/k` | Navigate up/down |
| `h/l` | Collapse/expand |
| `Enter` | Select/expand |
| `/` | Open filter bar |
| `n` | Next page |
| `p` | Previous page |
| `r` | Refresh data |
| `Tab` | Switch panes |
| `?` | Show help |
| `q` | Quit/back |

## Web UI Quick Start (Coming Soon)

::: info Development Status
The web interface is currently under active development (Phase 5). Basic functionality works, but features may be incomplete. For production use, please use the TUI interface with `kassie tui`.
:::

<!-- Commented out until web UI is fully implemented
```bash
kassie web
```

The web UI will open automatically. You'll see:

1. **Connection Page**: Select your profile and connect
2. **Explorer View**: Similar layout to TUI with resizable panels
3. **Filter Bar**: Type WHERE clauses with autocomplete
4. **Inspector Panel**: Click any row to see detailed JSON view
-->

## Next Steps

Now that you're up and running:

- [Configuration Guide](/guide/configuration) - Learn about all configuration options
- [TUI Usage](/guide/tui-usage) - Master the terminal interface
- [Examples](/examples/) - See practical usage examples

## Need Help?

If you run into issues:
- Check the [Troubleshooting Guide](/guide/troubleshooting)
- Run with debug logging: `kassie tui --log-level debug`
- [Open an issue](https://github.com/kashifKhn/kassie/issues) on GitHub
