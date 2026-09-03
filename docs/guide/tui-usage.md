# TUI Usage

The Terminal User Interface (TUI) provides a fast, keyboard-driven way to explore your Cassandra and ScyllaDB databases.

## Launching the TUI

Start Kassie TUI:

```bash
kassie tui
```

With a specific profile:

```bash
kassie tui --profile production
```

Connect to a remote server:

```bash
kassie tui --server remote.example.com:50051
```

## Interface Overview

The TUI has multiple views that you navigate between:

### Connection View

When you first launch Kassie, you'll see the connection view:

```
┌──────────────────────────────────────────────────────┐
│                   KASSIE                             │
│          Database Explorer for Cassandra             │
├──────────────────────────────────────────────────────┤
│                                                      │
│   ► local           (127.0.0.1:9042)                │
│     staging         (staging-db:9042)                │
│     production      (prod-1:9042, prod-2:9042)       │
│                                                      │
├──────────────────────────────────────────────────────┤
│ j/k: Navigate  Enter: Connect  q: Quit              │
└──────────────────────────────────────────────────────┘
```

**Actions**:
- `j` or `↓`: Move down
- `k` or `↑`: Move up
- `Enter`: Connect to selected profile
- `q`: Quit Kassie

### Explorer View

After connecting, you'll see the main explorer interface with three panels:

```
┌─────────────────┬──────────────────────────────────────┬─────────────┐
│   KEYSPACES     │         DATA GRID                    │  INSPECTOR  │
├─────────────────┼──────────────────────────────────────┼─────────────┤
│ ► system        │  id         │ name      │ created   │             │
│   v app_data    │ ─────────── │ ───────── │ ─────────│             │
│     ► users     │  123...abc  │ John Doe  │ 2024-01-│             │
│     ► orders    │  456...def  │ Jane Smith│ 2024-01-│             │
│     ► products  │  789...ghi  │ Bob Jones │ 2024-01-│             │
│   system_auth   │                                      │             │
│   system_schema │                                      │             │
│                 │                                      │             │
├─────────────────┴──────────────────────────────────────┴─────────────┤
│ Connected: local@127.0.0.1  |  app_data.users  |  Page 1/5          │
└────────────────────────────────────────────────────────────────────┘
```

**Left Panel (Sidebar)**: Keyspace and table navigation  
**Center Panel (Data Grid)**: Table rows and columns  
**Right Panel (Inspector)**: Detailed view of selected row  
**Bottom**: Status bar with connection info and hints

## Navigation

### Sidebar Navigation

The sidebar shows a tree of keyspaces and tables:

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `l` / `→` / `Enter` | Expand keyspace or select table |
| `h` / `←` | Collapse keyspace |
| `/` | Search keyspaces/tables |
| `Esc` | Clear search |

**Example workflow**:
1. Press `j` to navigate to `app_data` keyspace
2. Press `l` or `Enter` to expand it
3. Press `j` to move to `users` table
4. Press `Enter` to load table data

### Data Grid Navigation

When viewing table data:

| Key | Action |
|-----|--------|
| `j` / `↓` | Move to next row |
| `k` / `↑` | Move to previous row |
| `h` / `←` | Scroll left |
| `l` / `→` | Scroll right |
| `Enter` | View row details in inspector |
| `n` | Next page |
| `p` | Previous page |
| `r` | Refresh data |
| `/` | Open filter bar |

**Scrolling**:
- Use `h/l` to scroll horizontally through columns
- Use `j/k` to scroll vertically through rows
- Large tables are paginated automatically

### Switching Panels

| Key | Action |
|-----|--------|
| `Tab` | Switch to next panel |
| `Shift+Tab` | Switch to previous panel |

The active panel is highlighted with a colored border.

## Filtering Data

Press `/` to open the filter bar at the bottom:

```
┌────────────────────────────────────────────────────────┐
│ Filter: id = '550e8400-e29b-41d4-a716-446655440000'   │
└────────────────────────────────────────────────────────┘
```

**Filter Examples**:

```cql
# Filter by partition key
id = '550e8400-e29b-41d4-a716-446655440000'

# Filter by clustering key range
created_at > '2024-01-01' AND created_at < '2024-02-01'

# Filter with IN clause
status IN ('active', 'pending')

# Combine filters
user_id = 123 AND status = 'active'
```

**Keys in filter bar**:
- `Enter`: Apply filter
- `Esc`: Cancel and close filter bar
- `Ctrl+U`: Clear filter input

::: tip
Kassie validates your filter syntax before sending it to the database. Invalid filters will show an error.
:::

## Running Ad-hoc Queries

Press `Ctrl+O` to open the query bar and run any read-only SELECT:

```
┌────────────────────────────────────────────────────────┐
│ » SELECT id, email FROM app_data.users LIMIT 50        │
└────────────────────────────────────────────────────────┘
```

- `Enter` runs the query; results render in the data grid with full
  cursor paging (`n` for next page)
- `Esc` cancels
- Only `SELECT` statements are allowed — writes, DDL, `USE`, and
  multi-statement input are rejected before hitting the database

## Query History and Saved Queries

- `Ctrl+Y` — browse recent queries (newest first, 100-entry ring per
  profile; successful runs are recorded automatically)
- `Ctrl+P` — browse saved queries; press `d` on an entry to delete it

In both lists: `j/k` to navigate, `Enter` to run the selected query,
`Esc` to close.

::: tip
History and saved queries persist per profile in `~/.config/kassie/queries.json`, so they survive restarts.
:::

## Exporting Data

- `Ctrl+E` — export the currently loaded rows to a JSON file
  (`~/kassie-<keyspace>-<table>-<timestamp>.json`)
- `E` — stream a **full table** export server-side to
  `~/kassie-<keyspace>-<table>-<timestamp>-full.csv`. The server pages
  through Cassandra and streams chunks, so this works for tables far
  larger than memory.

## Inspector Panel

The inspector panel shows detailed row information with multiple viewing modes.

### Display Modes

Press `t` to cycle between display modes:

1. **Table Mode**: Type-aware layout with keys, CQL types, and values
   ```
   id          │ uuid        │ "550e8400-e29b-41d4-a716-446655440000"
   name        │ text        │ "John Doe"
   attrs       │ map<varchar, int> │ {"a": 1}
   payload     │ blob        │ 00000000  de ad be ef  |....|
   ```

   Blobs render as a hex dump with an ascii gutter; collections
   (maps/lists/sets/tuples) pretty-print as JSON. The type column is
   hidden on narrow panels.

2. **JSON Mode**: Pretty-printed JSON with syntax highlighting;
   collection cells are unwrapped into real JSON objects
   ```json
   {
     "id": "550e8400-e29b-41d4-a716-446655440000",
     "name": "John Doe",
     "email": "john@example.com",
     "created_at": "2024-01-15T10:30:00Z",
     "attrs": { "a": 1 }
   }
   ```

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down one line |
| `k` / `↑` | Scroll up one line |
| `h` / `←` | Scroll left (for long values) |
| `l` / `→` | Scroll right (for long values) |
| `d` | Page down (20 lines) |
| `u` | Page up (20 lines) |
| `[` | Navigate to previous row |
| `]` | Navigate to next row |
| `t` | Toggle display mode (Table/JSON) |
| `i` | Toggle fullscreen inspector mode |
| `Ctrl+C` | Copy content to clipboard |

### Horizontal Scrolling

For rows with long values (URLs, large IDs, JSON objects):

1. Focus the inspector panel (`Tab` or `Ctrl+I`)
2. Press `h` to scroll left, `l` to scroll right
3. The key column stays fixed while values scroll
4. A scroll indicator `[→ N]` shows your position

**Example**: Viewing a row with a long URL:
```
id                   │ "441f1d36-32bc-4259-b4de-342a30a3e142"
image_url            │ "https://example.com/very/long/path/to/image..."
```
Press `l` to scroll right:
```
id                   │ "2bc-4259-b4de-342a30a3e142"
image_url            │ .com/very/long/path/to/image/file.png"
```

### Fullscreen Mode

Press `i` when the inspector is focused to toggle fullscreen mode:

- **Normal mode**: Inspector shares space with grid and sidebar
- **Fullscreen mode**: Inspector takes the full screen for better viewing

This is useful for:
- Inspecting rows with many columns
- Viewing large JSON objects
- Reading long text values

Press `i` again to exit fullscreen.

### Row Navigation

Browse through rows without leaving the inspector:

1. View a row in the inspector (`Enter` from grid)
2. Press `]` to move to the next row
3. Press `[` to move to the previous row
4. The inspector updates automatically

This is much faster than switching back to the grid for each row.

### Clipboard Support

Press `Ctrl+C` to copy the current row data to clipboard:

- **Table mode**: Copies as formatted table
- **JSON mode**: Copies as JSON object

Requires clipboard utilities:
- **Linux**: `xclip`, `xsel`, or `wl-copy` (Wayland)
- **macOS**: Built-in (`pbcopy`)
- **Windows**: Built-in (`clip`)

::: tip Navigation Workflow
**Efficient data browsing**:
1. Select a row (`Enter` in grid)
2. Press `i` for fullscreen inspector
3. Use `]` to browse through rows
4. Use `h/l` to see long values
5. Press `i` to exit fullscreen
:::

## Pagination

Large tables are automatically paginated:

| Key | Action |
|-----|--------|
| `n` | Next page |
| `p` | Previous page |
| `g` | Go to first page |
| `G` | Go to last page |

The status bar shows current page:
```
Page 2/10  |  Rows: 50-100 of 1,000
```

::: info
Kassie uses Cassandra's paging state tokens for efficient pagination. No data is cached in memory.
:::

## Keyboard Shortcuts Reference

### Global

| Key | Action |
|-----|--------|
| `?` | Show help screen |
| `q` | Quit or go back |
| `Esc` | Cancel current action |
| `Tab` | Switch panels |
| `Ctrl+C` | Force quit |

### Sidebar

| Key | Action |
|-----|--------|
| `j/k` or `↓/↑` | Navigate up/down |
| `h/l` or `←/→` | Collapse/expand |
| `Enter` | Select table |
| `/` | Search |

### Data Grid

| Key | Action |
|-----|--------|
| `j/k` or `↓/↑` | Navigate rows |
| `h/l` or `←/→` | Scroll columns |
| `Enter` | View row details |
| `n/p` | Next/previous page |
| `g/G` | First/last page |
| `r` | Refresh data |
| `/` | Filter |

### Inspector

| Key | Action |
|-----|--------|
| `j/k` or `↓/↑` | Scroll vertically |
| `h/l` or `←/→` | Scroll horizontally |
| `[` / `]` | Previous/next row |
| `i` | Toggle fullscreen |
| `m` | Switch display mode |
| `Ctrl+C` | Copy to clipboard |
| `Esc` | Close |

## Themes

Kassie supports color themes. Configure in `config.json`:

```json
{
  "clients": {
    "tui": {
      "theme": "default"
    }
  }
}
```

::: info Theme Development
Currently only the `default` theme is fully implemented. The color scheme is optimized for both light and dark terminal backgrounds.

**Coming Soon:**
- `dracula`: Dark purple theme
- `nord`: Arctic-inspired colors
- `gruvbox`: Retro groove theme

Follow [GitHub issue #XX](https://github.com/kashifKhn/kassie/issues) for theme development progress.
:::

## Vim Mode

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

When enabled:
- `hjkl` for navigation
- `gg` / `G` for first/last
- `:q` to quit
- `/` for search (already enabled by default)

## Tips and Tricks

### Quick Navigation

1. **Jump to system tables**: Press `j` twice from the top to reach `system_schema`
2. **Fast filtering**: Press `/` and start typing immediately
3. **Inspect without selecting**: Some terminals support mouse clicks

### Efficient Workflows

**Exploring a new cluster**:
1. Connect and expand `system_schema`
2. Select `tables` to see all tables
3. Filter by keyspace: `keyspace_name = 'app_data'`

**Finding a specific record**:
1. Navigate to table
2. Press `/` and filter by primary key
3. Press `Enter` to view details

**Reviewing recent data**:
1. Select table with timestamp column
2. Filter: `created_at > '2024-01-01'`
3. Use `n/p` to page through results

### Performance Tips

- Use filters to reduce dataset size
- Smaller page sizes load faster (configure in `defaults.page_size`)
- Close inspector when not needed (press `Esc`)

## Troubleshooting

### TUI is slow

- Reduce `page_size` in config
- Use filters to limit data
- Check network latency to database

### Characters look broken

Your terminal may not support Unicode. Try:
```bash
export LANG=en_US.UTF-8
kassie tui
```

**Tested terminals**:
- **macOS**: iTerm2 (recommended), Terminal.app, Alacritty, Kitty
- **Linux**: GNOME Terminal, Konsole, Alacritty, Kitty, Tilix
- **Windows**: Windows Terminal (recommended), ConEmu, Mintty

See [Compatibility Guide](/guide/compatibility#terminal-compatibility) for full terminal compatibility matrix.

### Colors are wrong

Some terminals have limited color support. Try:
```bash
export TERM=xterm-256color
kassie tui
```

### Mouse doesn't work

Mouse support depends on your terminal. Keyboard navigation always works.

## Next Steps

- [Configuration](/guide/configuration) - Customize your setup
- [Keyboard Shortcuts Reference](/reference/keyboard-shortcuts) - Complete shortcut list
- [Examples](/examples/) - See practical examples
