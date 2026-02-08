---
layout: home

hero:
  name: Kassie
  text: Database Explorer for Cassandra & ScyllaDB
  tagline: Modern terminal and web interface for exploring your Cassandra and ScyllaDB clusters
  image:
    src: /logo.svg
    alt: Kassie
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/KashifKhn/kassie

features:
  - icon: 🖥️
    title: Dual Interface
    details: Choose between a fast terminal UI (TUI) or a modern web interface. Both are first-class citizens with full feature parity.
  
  - icon: 📦
    title: Single Binary
    details: No dependencies, no runtime requirements. One binary contains the server, TUI client, and web UI. Just download and run.
  
  - icon: ⚡
    title: Fast & Lightweight
    details: Built in Go for maximum performance. TUI launches in <100ms, web UI in <500ms. Minimal resource usage (<50MB RAM).
  
  - icon: 🔐
    title: Secure Authentication
    details: JWT-based authentication with access and refresh tokens. Credentials never exposed to clients. Session management built-in.
  
  - icon: ⌨️
    title: Keyboard-Driven
    details: Vim-like navigation in TUI. Full keyboard shortcuts in web UI. Designed for terminal enthusiasts and power users.
  
  - icon: 🎨
    title: Beautiful Design
    details: Polished TUI with multiple themes. Modern web UI with dark mode. Syntax highlighting and JSON inspection built-in.

  - icon: 🔍
    title: Smart Filtering
    details: Apply WHERE clauses with syntax validation. Filter by partition keys, clustering keys, and more. Query history included.
  
  - icon: 📄
    title: Pagination Made Easy
    details: Seamless pagination through large datasets. Cursor-based navigation. No memory issues with million-row tables.
  
  - icon: 🔄
    title: Real-time Schema
    details: Browse keyspaces, tables, and columns instantly. View partition keys, clustering keys, and indexes. Schema caching for performance.
---

## Quick Start

Get up and running in 30 seconds:

::: code-group

```bash [Homebrew]
brew tap KashifKhn/kassie
brew install kassie
kassie tui
```

```bash [Go Install]
go install github.com/KashifKhn/kassie@latest
kassie tui
```

```bash [Curl]
curl -sSL https://kassie.dev/install.sh | bash
kassie tui
```

```bash [Docker]
docker run -it ghcr.io/kashifkhn/kassie tui
```

:::

## Why Kassie?

### The Problem

Working with Cassandra and ScyllaDB traditionally means:
- Typing repetitive `DESCRIBE` commands in `cqlsh`
- Manually paginating through results
- No visual hierarchy or navigation
- Risk of accidental data mutations
- Poor exploration experience

### The Solution

Kassie provides a modern, safe way to explore your data:
- **Visual Navigation**: Tree-view sidebar with all keyspaces and tables
- **Read-Safety First**: Optimized for browsing and observing data
- **Multiple Interfaces**: Terminal UI for SSH sessions, Web UI for teams
- **Single Binary**: No installation hassles, works everywhere

## Comparison

| Feature | Kassie | cqlsh | Other GUIs |
|---------|--------|-------|------------|
| Installation | Single binary | Python + deps | Complex setup |
| Interface | TUI + Web | CLI only | GUI only |
| Navigation | Visual tree | Manual queries | Varies |
| Performance | Fast (<100ms) | Slow | Varies |
| Remote Access | Built-in server | SSH only | Requires setup |
| Keyboard-Driven | ✓ | Partial | ✗ |
| Filtering | Smart validation | Manual CQL | Varies |
| Authentication | JWT tokens | Password only | Varies |

## Use Cases

### Development
Quickly inspect data during development. View schema changes. Test queries with instant feedback.

### DevOps
SSH into production servers and use TUI to investigate issues. No GUI required. Fast and efficient.

### Team Collaboration
Run Kassie server, share the web UI link. Everyone can explore the same cluster with proper authentication.

### Data Exploration
Browse large tables with pagination. Filter by keys. Inspect complex data types (maps, sets, UDTs) in JSON format.

## What's Next?

<div class="vp-doc" style="margin-top: 2rem;">

- [Getting Started →](/guide/getting-started) - 5-minute tutorial
- [Installation →](/guide/installation) - Detailed installation guide
- [TUI Usage →](/guide/tui-usage) - Learn the terminal interface
- [Architecture →](/architecture/) - Understand how it works

</div>
