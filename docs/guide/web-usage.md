# Web Usage

The Web UI provides a modern, browser-based interface for exploring your Cassandra and ScyllaDB databases with a responsive, feature-rich experience.

## Launching the Web UI

Start Kassie Web:

```bash
kassie web
```

The web interface will open automatically at `http://localhost:9042`.

**Custom port**:
```bash
kassie web --port 3000
```

**Disable auto-open browser**:
```bash
kassie web --no-browser
```

## Interface Overview

The Web UI features a responsive, resizable three-panel layout optimized for database exploration.

### Login Page

When you first access Kassie, you'll see the profile selection page.

**Features**:
- Clean profile cards with connection details
- Host and port display
- Default keyspace indication
- One-click connection
- Loading states during authentication

**Connecting**:
1. Review available profiles
2. Click on a profile card
3. Wait for authentication
4. Automatically redirected to explorer

**Authentication**:
- JWT-based secure authentication
- Automatic token refresh
- Session persistence
- Secure logout

### Explorer Page

After connecting, the main explorer interface provides full database access.

**Layout Panels**:
- **Left Sidebar**: Hierarchical keyspace and table navigation
- **Center Panel**: Data grid with filter bar and pagination
- **Right Panel**: Inspector for detailed row examination
- **Header**: Theme toggle, panel controls, profile info, logout

All panels are resizable via drag handles between sections.

## Navigation

### Sidebar

The sidebar displays an expandable tree of keyspaces and tables.

**Features**:
- Collapsible keyspace sections
- Lazy-loaded table lists (fetched on expansion)
- Table row count estimates
- Visual indicators for partition keys (PK) and clustering keys (CK)
- Responsive loading states
- Panel toggle from header

**Actions**:
- Click keyspace name to expand/collapse
- Click table name to load data in grid
- Selected table highlighted
- Automatic state synchronization

**Indicators**:
- 🔑 **PK** badge: Partition key column
- 🔗 **CK** badge: Clustering key column
- Row counts formatted (K for thousands, M for millions)

### Data Grid

The center panel displays table data in a high-performance virtualized grid.

**Features**:
- Virtual scrolling (handles large datasets efficiently)
- Column type display in headers
- Key column badges (PK/CK)
- Row selection with click
- Cursor-based pagination
- Empty state messaging
- Error state handling

**Column Headers**:
Each column header displays:
- Column name
- Data type (text, int, uuid, etc.)
- Key badges (PK for partition key, CK for clustering key)

**Row Operations**:
- Click any row to view details in inspector
- Selected row highlighted with accent color
- Hover effect for better visibility

**Data Display**:
- Proper formatting for all CQL types
- NULL values clearly indicated
- Truncated long values with ellipsis
- Byte arrays shown as `<bytes>`
- Timestamps and UUIDs formatted

### Inspector Panel

The right panel shows detailed views of selected rows.

**View Modes**:
1. **Key-Value View**: Structured field-by-field display
2. **JSON View**: Syntax-highlighted JSON representation

**Features**:
- Collapsible nested objects/collections
- Color-coded JSON syntax
- Clean, readable formatting
- Toggle between views
- Empty state when no row selected
- Resizable panel

**JSON View**:
- Syntax highlighting for types
- Collapsible arrays and objects
- Copy-friendly formatting
- Dark mode optimized colors

## Filtering Data

Use the filter bar above the data grid to apply CQL WHERE clauses.

**Features**:
- Live filtering as you type WHERE clauses
- Clear button to remove filters
- Automatic query invalidation
- Error feedback for invalid syntax

**Example Filters**:

```cql
# Simple equality
id = '550e8400-e29b-41d4-a716-446655440000'

# Range query
created_at > '2024-01-01' AND created_at < '2024-02-01'

# IN clause
status IN ('active', 'pending', 'completed')

# Multiple conditions
user_id = 123 AND status = 'active'
```

**Using the Filter Bar**:
1. Type your WHERE clause in the search input
2. Click "Filter" button or press Enter
3. Data grid automatically refetches with filter
4. Click X icon to clear filter

**Important Notes**:
- Only WHERE clause content (no "WHERE" keyword needed)
- Must follow CQL syntax rules
- Partition key filters recommended for performance
- Filter persists during session

## Pagination

Navigate large datasets with smart cursor-based pagination.

**Features**:
- Next/Previous page buttons
- Cursor-based navigation (Cassandra-native)
- Automatic page state management
- Loading indicators during fetch
- "More data available" indicator

**Controls**:
- **Previous**: Navigate to prior page (appears after loading multiple pages)
- **Next**: Fetch next page using cursor (appears when more data available)
- **Row counter**: Shows total rows loaded

**Pagination Flow**:
```
Initial load → 100 rows + cursor
Click Next → Loads 100 more rows + new cursor
Click Previous → Returns to previous page state
```

**Performance**:
- Only visible rows rendered (virtualized)
- Cursor-based pagination (no offset scanning)
- Efficient for tables with millions of rows

## Theme Switching

Toggle between light, dark, and system themes.

**Theme Options**:
- **Light**: Bright background, dark text
- **Dark**: Dark background, light text  
- **System**: Automatically matches OS preference

**How to Switch**:
1. Click theme icon in header (☀️ Sun / 🌙 Moon / 💻 Monitor)
2. Cycles: Light → Dark → System → Light...
3. Preference saved in browser localStorage

**Theme Persistence**:
- Remembered across sessions
- Syncs with system preferences in "System" mode
- Instant switching without reload

## Notifications

The app uses toast notifications for user feedback.

**Toast Types**:
- ✅ **Success**: Green - successful operations (e.g., "Connected to local")
- ❌ **Error**: Red - failures and errors
- ℹ️ **Info**: Blue - informational messages
- ⚠️ **Warning**: Yellow - warnings

**Behavior**:
- Auto-dismiss after 3 seconds
- Manual dismiss with X button
- Stacked in bottom-right corner
- Non-blocking interface

**Common Notifications**:
- Login success/failure
- Connection errors
- Query execution feedback
- Logout confirmation

## Error Handling

Comprehensive error states throughout the application.

**Error Boundary**:
- Catches unexpected React errors
- User-friendly error page
- Reload button to recover
- Error logged to console

**Component Error States**:
- **Loading**: Animated spinner with message
- **Error**: Alert icon with error details
- **Empty**: Contextual empty state message

**Data Grid Errors**:
- Query failures shown with error message
- Network errors displayed clearly
- Retry functionality available

## Responsive Design

The Web UI adapts to different screen sizes.

### Desktop (>1024px)

Full three-panel layout with all features visible.

### Tablet (768-1024px)

- Collapsible sidebar
- Touch-friendly controls
- Optimized spacing

### Mobile (<768px)

- Single panel focus
- Drawer navigation
- Mobile-optimized data grid
- Full-screen inspector

## Browser Support

Kassie Web UI supports modern browsers:

| Browser | Minimum Version |
|---------|----------------|
| Chrome | 90+ |
| Firefox | 88+ |
| Safari | 14+ |
| Edge | 90+ |

::: warning
Internet Explorer is not supported.
:::

## Performance

### Virtual Scrolling

The data grid uses React Window for virtualization:
- Only visible rows rendered in DOM
- Handles tables with millions of rows
- Smooth scrolling performance
- Minimal memory footprint

### Caching

TanStack Query provides intelligent caching:
- Schema cached for 5 minutes
- Data cached until invalidated
- Background refetching
- Automatic cache invalidation on mutations

### Lazy Loading

Progressive data loading:
- Keyspaces loaded on mount
- Tables loaded per keyspace when expanded
- Data fetched per table when selected
- Pagination loads incrementally

## Technical Features

### State Management

- **Zustand** for UI state (theme, panel visibility, selection)
- **TanStack Query** for server state (data, schema)
- **localStorage** for persistence (auth, preferences)

### Type Safety

- Full TypeScript with strict mode
- Zod schema validation
- Type-safe API calls
- Zero `any` types

### Bundle Size

- **111.01 KB** gzipped (production build)
- Code splitting by route
- Optimized dependencies
- Tree-shaking enabled

### Security

- JWT-based authentication
- Automatic token refresh
- Secure storage (httpOnly where possible)
- CORS handling
- Protected routes

## Keyboard Shortcuts

While keyboard shortcuts are not yet fully implemented, the following work:

### Forms

| Key | Action |
|-----|--------|
| `Enter` | Submit filter / form |
| `Esc` | Clear/dismiss |

### Navigation

| Key | Action |
|-----|--------|
| Click | Select table / row |
| Scroll | Navigate data grid |

## Tips and Tricks

### Quick Navigation

1. Click profile to connect
2. Expand keyspace in sidebar
3. Click table to load
4. Apply filter if needed
5. Click row to inspect

### Efficient Filtering

For best performance:
- Include partition key in filter
- Use equality on partition key
- Avoid full table scans
- Test filters on small datasets first

### Panel Management

- Resize panels to your workflow
- Collapse sidebar for more grid space
- Toggle inspector when not needed
- Sizes persist in localStorage

### Theme Preference

- Use "System" to match OS dark mode
- Switch themes for different lighting
- Dark mode easier on eyes for long sessions

## Troubleshooting

### Page Won't Load

- Check browser console for errors
- Try incognito/private mode
- Clear browser cache
- Verify server is running

### Connection Failed

- Verify profile configuration
- Check Cassandra/ScyllaDB is running
- Verify network connectivity
- Check firewall rules

### Slow Data Loading

- Apply filters to reduce dataset
- Check network latency
- Verify database performance
- Reduce page size if needed

### Inspector Not Showing

- Check if panel collapsed (drag right border)
- Try selecting a row
- Refresh the page
- Check browser console

## Next Steps

- [Configuration](/guide/configuration) - Customize your setup
- [TUI Usage](/guide/tui-usage) - Use terminal interface
- [Troubleshooting](/guide/troubleshooting) - Fix common issues
