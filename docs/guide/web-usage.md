# Web Usage

The Web UI provides a modern, browser-based interface for exploring your Cassandra and ScyllaDB databases.

## Launching the Web UI

Start Kassie Web:

```bash
kassie web
```

The web interface will open automatically at `http://localhost:8080`.

**Custom port**:
```bash
kassie web --port 3000
```

**Disable auto-open browser**:
```bash
kassie web --no-browser
```

## Interface Overview

The Web UI features a responsive, resizable layout:

### Connection Page

When you first access Kassie, you'll see the connection page.

**Features**:
- Profile cards showing connection details
- Connection status indicators
- Recent connections list
- Quick connect buttons

**Connecting**:
1. Select a profile card
2. Click "Connect"
3. Wait for connection confirmation
4. You'll be redirected to the explorer

### Explorer Page

After connecting, the main explorer interface appears.

**Layout**:
- **Left Sidebar**: Keyspace and table navigation
- **Center Panel**: Data grid with rows
- **Right Panel**: Inspector for row details
- **Top Bar**: Filter input and actions
- **Bottom Bar**: Status and pagination

All panels are resizable by dragging the dividers.

## Navigation

### Sidebar

The sidebar displays a tree of keyspaces and tables:

**Actions**:
- Click keyspace to expand/collapse
- Click table name to load data
- Use search box to filter keyspaces/tables
- Drag divider to resize sidebar

**Keyboard shortcuts**:
- `↑/↓`: Navigate up/down
- `←/→`: Collapse/expand
- `Enter`: Select table
- `/`: Focus search box

### Data Grid

The center panel shows table data in a virtualized grid:

**Features**:
- Sortable columns (click header)
- Resizable columns (drag column border)
- Row selection (click row)
- Virtual scrolling (handles millions of rows)
- Pagination controls at bottom

**Column operations**:
- **Click header**: Sort by column (asc/desc)
- **Drag column border**: Resize column
- **Double-click border**: Auto-fit column width

**Row selection**:
- Click any row to view details in inspector
- Selected row is highlighted

### Inspector Panel

The right panel shows detailed JSON view of selected row:

**Features**:
- Syntax-highlighted JSON
- Collapsible nested objects
- Copy to clipboard button
- Pretty-print formatting

**Operations**:
- **Click object**: Expand/collapse
- **Copy button**: Copy entire JSON
- **Close button**: Hide inspector

## Filtering Data

Use the filter bar at the top to apply WHERE clauses.

**Features**:
- Syntax validation as you type
- Autocomplete for column names
- Query history dropdown
- Quick filter templates

**Example filters**:

```cql
# Simple equality
id = '550e8400-e29b-41d4-a716-446655440000'

# Range query
created_at > '2024-01-01' AND created_at < '2024-02-01'

# IN clause
status IN ('active', 'pending', 'completed')

# Multiple conditions
user_id = 123 AND status = 'active' AND created_at > '2024-01-01'
```

**Using the filter bar**:
1. Click the filter input or press `/`
2. Type your WHERE clause
3. Press `Enter` or click "Apply"
4. Results update automatically

**Query history**:
- Click the clock icon to see recent queries
- Select a query to reapply it
- History persists in browser storage

## Pagination

Navigate large datasets with pagination controls.

**Controls**:
- **First**: Jump to first page
- **Previous**: Go back one page
- **Page input**: Jump to specific page
- **Next**: Go forward one page
- **Last**: Jump to last page

**Status display**:
```
Showing rows 51-100 of 1,247  |  Page 2 of 13
```

**Keyboard shortcuts**:
- `n`: Next page
- `p`: Previous page
- `g`: First page
- `G`: Last page

## URL State

The Web UI synchronizes state with the URL for easy sharing:

**URL format**:
```
http://localhost:8080/?keyspace=app_data&table=users&filter=status='active'&page=2
```

**Benefits**:
- Share specific views with team members
- Bookmark frequently used queries
- Browser back/forward works
- Refresh preserves state

**What's stored in URL**:
- Current keyspace
- Current table
- Active filter
- Current page number

## Responsive Design

The Web UI adapts to different screen sizes:

### Desktop (>1024px)

Full three-panel layout with all features visible.

### Tablet (768-1024px)

- Collapsible sidebar (hamburger menu)
- Stacked data grid and inspector
- Touch-friendly controls

### Mobile (<768px)

- Single panel view
- Navigation drawer for sidebar
- Full-screen data grid
- Modal inspector

**Mobile gestures**:
- Swipe right: Open sidebar drawer
- Swipe left: Close sidebar drawer
- Tap row: Open inspector modal
- Pull to refresh: Reload data

## Keyboard Shortcuts

The Web UI supports extensive keyboard navigation:

### Global

| Key | Action |
|-----|--------|
| `?` | Show keyboard shortcuts help |
| `/` | Focus filter bar |
| `Esc` | Cancel/close current action |
| `Ctrl+K` | Focus search |

### Navigation

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate rows |
| `←/→` | Scroll columns |
| `Enter` | View row details |
| `Tab` | Switch panels |

### Pagination

| Key | Action |
|-----|--------|
| `n` | Next page |
| `p` | Previous page |
| `g` | First page |
| `G` | Last page |

### Actions

| Key | Action |
|-----|--------|
| `r` | Refresh data |
| `c` | Copy selected row |
| `i` | Toggle inspector |
| `s` | Toggle sidebar |

Press `?` in the Web UI to see the complete shortcuts list.

## Features

### Dark Mode

Toggle between light and dark themes:

- Click the theme icon in the top-right
- System preference auto-detected
- Preference saved in localStorage

### Connection Management

**Switching profiles**:
1. Click profile name in top-left
2. Select "Change Connection"
3. Choose new profile
4. Confirm switch

**Logout**:
1. Click profile menu
2. Select "Logout"
3. Return to connection page

### Data Export

Export current view to JSON or CSV:

1. Click "Export" button
2. Choose format (JSON/CSV)
3. Optionally apply filters
4. Click "Download"

::: info
Export is limited to current result set. Use filters to reduce data size before exporting.
:::

### Clipboard Operations

**Copy row**:
- Right-click row → Copy JSON
- Or press `c` with row selected

**Copy cell**:
- Click cell → Click copy icon
- Or double-click cell to select text

### Connection Status

The status bar shows real-time connection info:

```
Connected to local@127.0.0.1:9042  |  app_data.users  |  Last updated: 2 seconds ago
```

**Indicators**:
- 🟢 Green: Connected
- 🟡 Yellow: Connecting
- 🔴 Red: Disconnected

If connection is lost, Kassie attempts automatic reconnection.

## Customization

### Panel Sizes

Resize panels by dragging dividers:
- Drag left divider to resize sidebar
- Drag right divider to resize inspector
- Sizes saved in localStorage

### Column Widths

Adjust column widths in data grid:
- Drag column borders
- Double-click to auto-fit
- Widths persist per table

### Preferences

Configure preferences via profile menu:

**Available settings**:
- Theme (light/dark/auto)
- Auto-open inspector
- Default page size
- Sidebar collapsed by default
- Show line numbers in JSON

## Performance

### Virtual Scrolling

The data grid uses virtualization for performance:
- Only visible rows are rendered
- Handles tables with millions of rows
- Smooth scrolling even with 1000+ columns

### Caching

Smart caching reduces database queries:
- Schema cached for 5 minutes
- Navigation state cached
- Query results cached until refresh

### Lazy Loading

Data loads progressively:
- Initial page loads immediately
- Subsequent pages fetch on demand
- No upfront cost for large datasets

## Accessibility

The Web UI follows WCAG 2.1 AA standards:

**Features**:
- Full keyboard navigation
- Screen reader support
- ARIA labels on interactive elements
- Focus management
- High contrast mode
- Reduced motion support

**Screen reader tested with**:
- NVDA (Windows)
- JAWS (Windows)
- VoiceOver (macOS)

## Browser Support

Kassie Web UI supports:

| Browser | Minimum Version |
|---------|----------------|
| Chrome | 90+ |
| Firefox | 88+ |
| Safari | 14+ |
| Edge | 90+ |

::: warning
Internet Explorer is not supported.
:::

## Tips and Tricks

### Quick Filters

Create bookmarkable URLs for common filters:

```
# Active users
http://localhost:8080/?keyspace=app&table=users&filter=status='active'

# Recent orders
http://localhost:8080/?keyspace=app&table=orders&filter=created_at>'2024-01-01'
```

### Keyboard-First Workflow

1. Launch Kassie: `kassie web`
2. Press `/` to search keyspaces
3. Press `Enter` to select table
4. Press `/` to add filter
5. Use `n/p` to navigate pages
6. Press `Enter` to inspect rows

### Team Sharing

Run Kassie server for team access:

```bash
# Start server
kassie server --http-port 8080 --host 0.0.0.0

# Share URL
http://your-server:8080
```

Everyone can connect with their own profiles.

## Troubleshooting

### Page won't load

- Check browser console for errors
- Try incognito/private mode
- Clear browser cache and cookies

### WebSocket connection failed

Kassie uses gRPC-Web which requires HTTP/2:
- Ensure server supports HTTP/2
- Check for proxy/firewall blocking connections

### Slow data loading

- Reduce page size in config
- Apply filters to limit data
- Check network latency
- Enable browser's hardware acceleration

### Inspector not showing

- Try resizing panels (inspector might be collapsed)
- Check browser console for JavaScript errors
- Refresh the page

## Next Steps

- [Configuration](/guide/configuration) - Customize your setup
- [API Reference](/reference/api) - Use the REST API
- [Examples](/examples/) - See practical examples
