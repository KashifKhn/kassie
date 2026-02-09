# Keyboard Shortcuts

Complete reference for all keyboard shortcuts in Kassie.

## TUI Shortcuts

### Global (All Views)

| Key | Action |
|-----|--------|
| `?` | Show help screen |
| `q` | Quit or go back to previous view |
| `Esc` | Cancel current action or close dialog |
| `Ctrl+C` | Force quit Kassie |

### Connection View

| Key | Action |
|-----|--------|
| `j` or `↓` | Move down in profile list |
| `k` or `↑` | Move up in profile list |
| `Enter` | Connect to selected profile |
| `q` | Quit Kassie |

### Sidebar (Explorer View)

| Key | Action |
|-----|--------|
| `j` or `↓` | Navigate down |
| `k` or `↑` | Navigate up |
| `h` or `←` | Collapse keyspace/go up level |
| `l` or `→` | Expand keyspace/enter |
| `Enter` | Select table or expand keyspace |
| `/` | Focus search/filter input |
| `Ctrl+F` | Focus search/filter input |
| `Esc` | Clear search |

### Data Grid (Explorer View)

| Key | Action |
|-----|--------|
| `j` or `↓` | Move to next row |
| `k` or `↑` | Move to previous row |
| `h` or `←` | Scroll left (previous columns) |
| `l` or `→` | Scroll right (next columns) |
| `Enter` | View row details in inspector |
| `g` | Go to first row |
| `G` | Go to last row |
| `n` | Next page |
| `p` | Previous page |
| `r` | Refresh data |
| `/` | Open filter bar |
| `Ctrl+F` | Focus search input |

### Filter Bar

| Key | Action |
|-----|--------|
| `Enter` | Apply filter |
| `Esc` | Cancel and close filter bar |
| `Ctrl+U` | Clear filter input |
| `↑` | Previous filter from history |
| `↓` | Next filter from history |

### Inspector Panel

| Key | Action |
|-----|--------|
| `j` or `↓` | Scroll down one line |
| `k` or `↑` | Scroll up one line |
| `d` | Page down (20 lines) |
| `u` | Page up (20 lines) |
| `t` | Toggle display mode (table/JSON/formatted) |
| `Ctrl+C` | Copy content to clipboard |

### Panel Navigation

| Key | Action |
|-----|--------|
| `Tab` | Switch to next panel (Sidebar → Grid → Inspector) |
| `Shift+Tab` | Switch to previous panel |
| `Ctrl+H` | Focus sidebar panel |
| `Ctrl+L` | Focus grid panel |
| `Ctrl+I` | Focus inspector panel |
| `Ctrl+B` | Toggle view mode (Full → No Sidebar → Grid Only → Full) |
| `Ctrl+F` | Activate search in current panel |

### Help View

| Key | Action |
|-----|--------|
| `j` or `↓` | Scroll down one line |
| `k` or `↑` | Scroll up one line |
| `d` | Page down (10 lines) |
| `u` | Page up (10 lines) |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `q` or `Esc` | Close help and return to explorer |

## Web UI Shortcuts

### Global

| Key | Action |
|-----|--------|
| `?` | Show keyboard shortcuts help |
| `/` | Focus filter bar |
| `Esc` | Cancel/close current action |
| `Ctrl+K` or `Cmd+K` | Focus search |

### Navigation

| Key | Action |
|-----|--------|
| `↑` | Navigate up in current context |
| `↓` | Navigate down in current context |
| `←` | Navigate left / collapse |
| `→` | Navigate right / expand |
| `Enter` | Select/confirm action |
| `Tab` | Switch between panels |

### Sidebar

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate keyspaces/tables |
| `←/→` | Collapse/expand keyspace |
| `Enter` | Select table |
| `/` | Focus search box |
| `s` | Toggle sidebar visibility |

### Data Grid

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate rows |
| `←/→` | Scroll columns |
| `Enter` | View row details |
| `n` | Next page |
| `p` | Previous page |
| `g` | First page |
| `G` | Last page |
| `r` | Refresh data |
| `c` | Copy selected row |

### Inspector

| Key | Action |
|-----|--------|
| `↑/↓` | Scroll |
| `Enter` | Expand/collapse node |
| `Esc` | Close inspector |
| `i` | Toggle inspector visibility |
| `c` | Copy JSON |

### Filter Bar

| Key | Action |
|-----|--------|
| `Enter` | Apply filter |
| `Esc` | Cancel filter |
| `↑/↓` | Navigate filter history |
| `Ctrl+U` or `Cmd+U` | Clear input |

## Vim Mode (TUI Only)

Enable Vim mode in config:
```json
{
  "clients": {
    "tui": {
      "vim_mode": true
    }
  }
}
```

### Additional Vim Bindings

| Key | Action |
|-----|--------|
| `h` | Move left (replaces `←`) |
| `j` | Move down (replaces `↓`) |
| `k` | Move up (replaces `↑`) |
| `l` | Move right (replaces `→`) |
| `gg` | Go to first item |
| `G` | Go to last item |
| `0` | Go to line start |
| `$` | Go to line end |
| `w` | Next word |
| `b` | Previous word |
| `:q` | Quit |
| `/` | Search (already standard) |
| `n` | Next search result |
| `N` | Previous search result |

## Mouse Support

### TUI (Terminal Dependent)

Many modern terminals support mouse operations:

| Action | Result |
|--------|--------|
| Click keyspace | Expand/collapse |
| Click table | Select and load data |
| Click row | Select row |
| Double-click row | View in inspector |
| Scroll wheel | Scroll current panel |

::: info
Mouse support depends on your terminal emulator. Keyboard navigation always works.
:::

### Web UI

Full mouse support:

| Action | Result |
|--------|--------|
| Click | Select |
| Double-click | Expand/collapse or view details |
| Drag | Resize panels |
| Scroll | Scroll content |
| Right-click | Context menu (where applicable) |

## Accessibility

### Screen Reader Support (Web UI)

The Web UI supports screen readers with proper ARIA labels:

| Key | Action |
|-----|--------|
| `Tab` | Navigate through interactive elements |
| `Shift+Tab` | Navigate backwards |
| `Enter` / `Space` | Activate element |

### Focus Navigation

| Key | Action |
|-----|--------|
| `Tab` | Next focusable element |
| `Shift+Tab` | Previous focusable element |
| `Esc` | Return focus to main area |

## Customization

### Remapping (Future Feature)

Currently, keyboard shortcuts are not customizable. This feature is planned for a future release.

## Tips

### Efficient Navigation

**TUI**:
- Use `j/k` for vertical, `h/l` for horizontal movement
- `g/G` to jump to first/last items quickly
- `/` for search is faster than scrolling

**Web UI**:
- Learn `?` to see all shortcuts
- Use `/` to quickly filter without clicking
- `Tab` to switch panels without mouse

### Common Workflows

**Quick table inspection**:
1. `j/k` to navigate to table
2. `Enter` to select
3. `/` to filter
4. `n/p` to page through

**Find specific record**:
1. Select table
2. `/` for filter
3. Type: `id = 'value'`
4. `Enter` to apply

**Browse multiple tables**:
1. `Tab` to sidebar
2. `j/k` to next table
3. `Enter` to load
4. Repeat

## Cheat Sheet

Print this cheat sheet for quick reference:

```
KASSIE KEYBOARD SHORTCUTS

Global:           Navigation:           Data Grid:
? - Help          j/k - Up/Down         Enter - Details
q - Quit          h/l - Left/Right      n/p - Next/Prev Page
Esc - Cancel      Tab - Switch Panel    r - Refresh
                  Ctrl+H/L/I - Focus    / - Filter

Panel Control:    Filter Bar:           Inspector:
Ctrl+B - Views    Enter - Apply         j/k - Scroll
Ctrl+F - Search   Esc - Cancel          d/u - Page Down/Up
                  ↑/↓ - History         t - Toggle Mode
                                        Ctrl+C - Copy
```

## Next Steps

- [TUI Usage](/guide/tui-usage) - Learn TUI workflows
- [Web Usage](/guide/web-usage) - Learn Web UI features
- [CLI Commands](/reference/cli-commands) - Command-line reference
