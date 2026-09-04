import { useEffect, useMemo, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Search, CornerDownLeft, Database, Terminal, PanelLeft, PanelRight, Sun, Moon, RefreshCw, LogOut } from 'lucide-react';
import { historyApi, queryKeys, schemaApi, sessionApi } from '@/api/queries';
import { useUiStore } from '@/stores/uiStore';
import { useAuthStore } from '@/stores/authStore';
import { useToastStore } from '@/stores/toastStore';

export interface PaletteCommand {
  id: string;
  label: string;
  group: string;
  icon: React.ReactNode;
  keywords?: string;
  run: () => void;
}

interface CommandPaletteProps {
  open: boolean;
  onClose: () => void;
  onRunQuery?: (cql: string) => void;
}

const RECENTS_KEY = 'kassie-palette-recents';

export function CommandPalette({ open, onClose, onRunQuery }: CommandPaletteProps) {
  const navigate = useNavigate();
  const { selectedKeyspace, setSelectedKeyspace, setSelectedTable, toggleSidebar, toggleInspector, setTheme, theme } = useUiStore();
  const { clearAuth } = useAuthStore();
  const { success } = useToastStore();

  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const keyspacesQuery = useQuery({
    queryKey: queryKeys.schema.keyspaces(),
    queryFn: schemaApi.listKeyspaces,
    enabled: open,
  });

  const savedQuery = useQuery({
    queryKey: queryKeys.history.saved(),
    queryFn: () => historyApi.listSavedQueries(),
    enabled: open,
  });

  useEffect(() => {
    if (open) {
      setQuery('');
      setSelected(0);
      setTimeout(() => inputRef.current?.focus(), 0);
    }
  }, [open]);

  const commands = useMemo<PaletteCommand[]>(() => {
    const cmds: PaletteCommand[] = [];

    cmds.push({
      id: 'action:refresh',
      label: 'Refresh current table',
      group: 'Actions',
      icon: <RefreshCw className="w-3.5 h-3.5" />,
      run: () => window.location.reload(),
    });

    cmds.push({
      id: 'action:toggle-sidebar',
      label: 'Toggle sidebar',
      group: 'Actions',
      icon: <PanelLeft className="w-3.5 h-3.5" />,
      run: () => toggleSidebar(),
    });

    cmds.push({
      id: 'action:toggle-inspector',
      label: 'Toggle inspector panel',
      group: 'Actions',
      icon: <PanelRight className="w-3.5 h-3.5" />,
      run: () => toggleInspector(),
    });

    cmds.push({
      id: 'action:toggle-theme',
      label: `Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`,
      group: 'Actions',
      icon: theme === 'dark' ? <Sun className="w-3.5 h-3.5" /> : <Moon className="w-3.5 h-3.5" />,
      run: () => setTheme(theme === 'dark' ? 'light' : 'dark'),
    });

    cmds.push({
      id: 'action:logout',
      label: 'Log out',
      group: 'Actions',
      icon: <LogOut className="w-3.5 h-3.5" />,
      run: () => {
        sessionApi.logout().finally(() => {
          clearAuth();
          navigate('/login');
        });
      },
    });

    for (const ks of keyspacesQuery.data?.keyspaces ?? []) {
      cmds.push({
        id: `ks:${ks.name}`,
        label: ks.name,
        group: 'Keyspaces',
        icon: <Database className="w-3.5 h-3.5" />,
        keywords: 'keyspace open browse',
        run: () => {
          setSelectedKeyspace(ks.name);
          setSelectedTable(null);
        },
      });
    }

    if (selectedKeyspace) {
      cmds.push({
        id: 'grid:open-query',
        label: 'Open CQL query editor',
        group: 'Actions',
        icon: <Terminal className="w-3.5 h-3.5" />,
        run: () => {
          const editor = document.querySelector<HTMLTextAreaElement>('textarea[placeholder^="SELECT"]');
          editor?.focus();
        },
      });
    }

    for (const sq of savedQuery.data ?? []) {
      cmds.push({
        id: `saved:${sq.name}`,
        label: `${sq.name} — ${sq.cql}`,
        group: 'Saved Queries',
        icon: <Terminal className="w-3.5 h-3.5" />,
        keywords: 'run query saved',
        run: () => {
          if (onRunQuery) {
            onRunQuery(sq.cql);
          } else {
            success(`Loaded: ${sq.name}`);
          }
        },
      });
    }

    return cmds;
  }, [keyspacesQuery.data, savedQuery.data, selectedKeyspace, theme, toggleSidebar, toggleInspector, setTheme, clearAuth, navigate, setSelectedKeyspace, setSelectedTable, onRunQuery, success]);

  const filtered = useMemo(() => {
    if (!query.trim()) return commands;

    const q = query.toLowerCase();
    const scored: Array<{ cmd: PaletteCommand; score: number }> = [];

    for (const cmd of commands) {
      const haystack = `${cmd.label} ${cmd.group} ${cmd.keywords ?? ''}`.toLowerCase();
      let score = 0;

      if (cmd.label.toLowerCase().startsWith(q)) score += 100;
      if (haystack.includes(q)) score += 50;

      let qi = 0;
      for (const ch of haystack) {
        if (qi < q.length && ch === q[qi]) qi++;
      }
      if (qi === q.length) score += 25;

      if (score > 0) scored.push({ cmd, score });
    }

    scoreRecents(scored);

    scored.sort((a, b) => b.score - a.score);
    return scored.map((s) => s.cmd);
  }, [commands, query]);

  useEffect(() => {
    setSelected(0);
  }, [query]);

  useEffect(() => {
    const item = listRef.current?.children[selected] as HTMLElement | undefined;
    item?.scrollIntoView({ block: 'nearest' });
  }, [selected]);

  if (!open) {
    return null;
  }

  const handleRun = (cmd: PaletteCommand) => {
    rememberRecent(cmd.id);
    cmd.run();
    onClose();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelected((prev) => Math.min(prev + 1, filtered.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelected((prev) => Math.max(prev - 1, 0));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const cmd = filtered[selected];
      if (cmd) handleRun(cmd);
    } else if (e.key === 'Escape') {
      onClose();
    }
  };

  let lastGroup = '';

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-24"
      style={{ background: 'rgba(0, 0, 0, 0.6)' }}
      onClick={onClose}
    >
      <div
        className="w-full max-w-xl rounded-xl glass animate-scale-in overflow-hidden"
        style={{
          background: 'var(--bg-elevated)',
          border: '1px solid var(--border-primary)',
          boxShadow: 'var(--shadow-glow, var(--shadow-lg))',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-3 px-4 py-3" style={{ borderBottom: '1px solid var(--border-primary)' }}>
          <Search className="w-4 h-4" style={{ color: 'var(--accent-primary)' }} />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search keyspaces, tables, queries, actions…"
            className="flex-1 bg-transparent outline-none text-sm font-mono"
            style={{ color: 'var(--text-primary)' }}
          />
          <span className="text-xs font-mono px-1.5 py-0.5 rounded" style={{ background: 'var(--bg-tertiary)', color: 'var(--text-tertiary)' }}>
            esc
          </span>
        </div>

        <div ref={listRef} className="max-h-80 overflow-auto py-1">
          {filtered.length === 0 && (
            <div className="px-4 py-6 text-center text-sm font-mono" style={{ color: 'var(--text-tertiary)' }}>
              No matching commands
            </div>
          )}
          {filtered.map((cmd, i) => {
            const showGroup = cmd.group !== lastGroup;
            lastGroup = cmd.group;
            return (
              <div key={cmd.id}>
                {showGroup && (
                  <div
                    className="px-4 pt-2 pb-1 text-[10px] font-mono font-bold uppercase tracking-wider"
                    style={{ color: 'var(--text-tertiary)' }}
                  >
                    {cmd.group}
                  </div>
                )}
                <div
                  className="flex items-center gap-3 px-4 py-2 text-sm font-mono cursor-pointer"
                  style={{ background: i === selected ? 'var(--accent-subtle)' : 'transparent' }}
                  onClick={() => handleRun(cmd)}
                  onMouseEnter={() => setSelected(i)}
                >
                  <span style={{ color: i === selected ? 'var(--accent-primary)' : 'var(--text-tertiary)' }}>
                    {cmd.icon}
                  </span>
                  <span
                    className="flex-1 truncate"
                    style={{ color: i === selected ? 'var(--text-primary)' : 'var(--text-secondary)' }}
                  >
                    {cmd.label}
                  </span>
                  {i === selected && (
                    <span className="flex items-center gap-1 text-xs" style={{ color: 'var(--text-tertiary)' }}>
                      <CornerDownLeft className="w-3 h-3" />
                    </span>
                  )}
                </div>
              </div>
            );
          })}
        </div>

        <div
          className="flex items-center gap-4 px-4 py-2 text-xs font-mono"
          style={{ borderTop: '1px solid var(--border-primary)', color: 'var(--text-tertiary)' }}
        >
          <span>↑↓ navigate</span>
          <span>↵ select</span>
          <span>esc close</span>
        </div>
      </div>
    </div>
  );
}

function scoreRecents(scored: Array<{ cmd: PaletteCommand; score: number }>) {
  const recents = getRecents();
  for (const entry of scored) {
    const idx = recents.indexOf(entry.cmd.id);
    if (idx >= 0) {
      entry.score += 30 - Math.min(idx, 10) * 2;
    }
  }
}

function getRecents(): string[] {
  try {
    return JSON.parse(localStorage.getItem(RECENTS_KEY) ?? '[]') as string[];
  } catch {
    return [];
  }
}

function rememberRecent(id: string) {
  const recents = getRecents().filter((r) => r !== id);
  recents.unshift(id);
  localStorage.setItem(RECENTS_KEY, JSON.stringify(recents.slice(0, 10)));
}
