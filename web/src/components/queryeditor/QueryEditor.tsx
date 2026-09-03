import { useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Play, Loader2, AlertCircle, ChevronDown, ChevronUp, Clock, Bookmark, Trash2 } from 'lucide-react';
import { dataApi, historyApi, queryKeys } from '@/api/queries';
import type { Row, CellValue, ExecuteQueryResponse, QueryHistoryEntry, SavedQuery } from '@/api/types';

interface QueryEditorProps {
  onResults?: (results: QueryResults | null) => void;
}

export interface QueryResults {
  rows: Row[];
  cursorId: string;
  hasMore: boolean;
}

export function QueryEditor({ onResults }: QueryEditorProps) {
  const [expanded, setExpanded] = useState(false);
  const [cql, setCql] = useState('');
  const [results, setResults] = useState<ExecuteQueryResponse | null>(null);
  const [allRows, setAllRows] = useState<Row[]>([]);

  const historyQuery = useQuery({
    queryKey: queryKeys.history.queries(),
    queryFn: () => historyApi.listHistory(20),
    enabled: expanded,
  });

  const savedQuery = useQuery({
    queryKey: queryKeys.history.saved(),
    queryFn: () => historyApi.listSavedQueries(),
    enabled: expanded,
  });

  const executeMutation = useMutation({
    mutationFn: dataApi.executeQuery,
    onSuccess: (data) => {
      setResults(data);
      setAllRows(data.rows);
      onResults?.({ rows: data.rows, cursorId: data.cursorId, hasMore: data.hasMore });
    },
  });

  const handleExecute = () => {
    if (!cql.trim()) return;
    executeMutation.mutate({ cql: cql.trim(), pageSize: 200 });
  };

  const handleNextPage = () => {
    if (!results || !results.cursorId) return;
    dataApi.getNextPage({ cursorId: results.cursorId }).then((data) => {
      setAllRows((prev) => [...prev, ...data.rows]);
      setResults((prev) => (prev ? { ...prev, cursorId: data.cursorId, hasMore: data.hasMore } : prev));
    });
  };

  const error = executeMutation.error instanceof Error
    ? { message: executeMutation.error.message }
    : null;

  return (
    <div
      className="flex-shrink-0"
      style={{
        borderBottom: '1px solid var(--border-primary)',
        background: 'var(--bg-secondary)',
      }}
    >
      <div
        className="flex items-center justify-between px-4 py-2 cursor-pointer select-none"
        onClick={() => setExpanded((prev) => !prev)}
      >
        <div className="flex items-center gap-2 text-sm font-mono font-bold" style={{ color: 'var(--text-primary)' }}>
          <Play className="w-4 h-4" style={{ color: 'var(--accent-primary)' }} />
          CQL Query
          {expanded ? (
            <ChevronUp className="w-4 h-4" style={{ color: 'var(--text-tertiary)' }} />
          ) : (
            <ChevronDown className="w-4 h-4" style={{ color: 'var(--text-tertiary)' }} />
          )}
        </div>
        <span className="text-xs font-mono" style={{ color: 'var(--text-tertiary)' }}>
          read-only SELECT
        </span>
      </div>

      {expanded && (
        <div className="px-4 pb-4 space-y-3 animate-fade-in">
          <div className="flex gap-2">
            <textarea
              value={cql}
              onChange={(e) => setCql(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                  handleExecute();
                }
              }}
              placeholder="SELECT * FROM keyspace.table WHERE ... LIMIT 100"
              rows={3}
              className="flex-1 px-3 py-2 text-sm font-mono rounded-lg resize-none outline-none"
              style={{
                background: 'var(--bg-primary)',
                border: '1px solid var(--border-primary)',
                color: 'var(--text-primary)',
              }}
            />
            <button
              onClick={handleExecute}
              disabled={!cql.trim() || executeMutation.isPending}
              className="flex items-center gap-2 px-4 self-stretch text-sm font-mono font-medium rounded-lg transition-all"
              style={{
                background: executeMutation.isPending ? 'var(--bg-tertiary)' : 'var(--accent-primary)',
                color: executeMutation.isPending ? 'var(--text-tertiary)' : 'var(--text-inverse)',
                cursor: executeMutation.isPending ? 'not-allowed' : 'pointer',
                opacity: !cql.trim() ? 0.5 : 1,
              }}
            >
              {executeMutation.isPending ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Play className="w-4 h-4" />
              )}
              Run
            </button>
          </div>

          <QueryPickers
            history={historyQuery.data ?? []}
            saved={savedQuery.data ?? []}
            onPick={(query) => {
              setCql(query);
              setResults(null);
              setAllRows([]);
            }}
          />

          {error && (
            <div
              className="flex items-center gap-2 px-3 py-2 text-xs font-mono rounded-lg"
              style={{
                background: 'var(--bg-primary)',
                border: '1px solid var(--error)',
                color: 'var(--error)',
              }}
            >
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              {error.message}
            </div>
          )}

          {results && (
            <QueryResultTable
              rows={allRows}
              hasMore={results.hasMore}
              totalFetched={results.totalFetched}
              onLoadMore={handleNextPage}
            />
          )}
        </div>
      )}
    </div>
  );
}

interface QueryPickersProps {
  history: QueryHistoryEntry[];
  saved: SavedQuery[];
  onPick: (cql: string) => void;
}

function QueryPickers({ history, saved, onPick }: QueryPickersProps) {
  const [showHistory, setShowHistory] = useState(false);
  const [showSaved, setShowSaved] = useState(false);
  const deleteMutation = useMutation({ mutationFn: historyApi.deleteSavedQuery });

  if (history.length === 0 && saved.length === 0) {
    return null;
  }

  return (
    <div className="flex gap-2">
      {history.length > 0 && (
        <div className="relative">
          <PickerToggle
            icon={<Clock className="w-3.5 h-3.5" />}
            label={`History (${history.length})`}
            active={showHistory}
            onClick={() => {
              setShowHistory((prev) => !prev);
              setShowSaved(false);
            }}
          />
          {showHistory && (
            <PickerList
              items={history.map((entry) => ({ key: entry.cql, label: entry.cql }))}
              onPick={(cql) => {
                onPick(cql);
                setShowHistory(false);
              }}
            />
          )}
        </div>
      )}
      {saved.length > 0 && (
        <div className="relative">
          <PickerToggle
            icon={<Bookmark className="w-3.5 h-3.5" />}
            label={`Saved (${saved.length})`}
            active={showSaved}
            onClick={() => {
              setShowSaved((prev) => !prev);
              setShowHistory(false);
            }}
          />
          {showSaved && (
            <PickerList
              items={saved.map((sq) => ({
                key: sq.name,
                label: `${sq.name}: ${sq.cql}`,
                onDelete: () => deleteMutation.mutate(sq.name),
              }))}
              onPick={(cql) => {
                onPick(cql);
                setShowSaved(false);
              }}
            />
          )}
        </div>
      )}
    </div>
  );
}

interface PickerItem {
  key: string;
  label: string;
  onDelete?: () => void;
}

function PickerToggle({
  icon,
  label,
  active,
  onClick,
}: {
  icon: React.ReactNode;
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono rounded-md transition-all"
      style={{
        background: active ? 'var(--accent-primary)' : 'var(--accent-subtle)',
        color: active ? 'var(--text-inverse)' : 'var(--accent-primary)',
      }}
    >
      {icon}
      {label}
    </button>
  );
}

function PickerList({ items, onPick }: { items: PickerItem[]; onPick: (cql: string) => void }) {
  return (
    <div
      className="absolute z-20 mt-1 max-w-md max-h-56 overflow-auto rounded-lg animate-fade-in"
      style={{
        background: 'var(--bg-elevated)',
        border: '1px solid var(--border-primary)',
        boxShadow: 'var(--shadow-lg)',
      }}
    >
      {items.map((item) => (
        <div
          key={item.key}
          className="flex items-center justify-between gap-2 px-3 py-2 text-xs font-mono cursor-pointer transition-all"
          style={{ borderBottom: '1px solid var(--border-primary)' }}
          onClick={() => onPick(item.label.includes(': ') ? item.label.split(': ').slice(1).join(': ') : item.label)}
          onMouseEnter={(e) => {
            e.currentTarget.style.background = 'var(--bg-secondary)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = 'transparent';
          }}
        >
          <span className="truncate" style={{ color: 'var(--text-primary)' }} title={item.label}>
            {item.label}
          </span>
          {item.onDelete && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                item.onDelete?.();
              }}
              style={{ color: 'var(--text-tertiary)' }}
              onMouseEnter={(e) => {
                e.currentTarget.style.color = 'var(--error)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.color = 'var(--text-tertiary)';
              }}
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
      ))}
    </div>
  );
}

interface QueryResultTableProps {
  rows: Row[];
  hasMore: boolean;
  totalFetched: number;
  onLoadMore: () => void;
}

function QueryResultTable({ rows, hasMore, totalFetched, onLoadMore }: QueryResultTableProps) {
  const columns = rows.length > 0 ? Object.keys(rows[0]?.cells ?? {}) : [];

  const cellText = (cell: CellValue | undefined): string => {
    if (!cell || cell.isNull) return 'NULL';
    if ('stringVal' in cell) return cell.stringVal;
    if ('intVal' in cell) return cell.intVal.toString();
    if ('doubleVal' in cell) return cell.doubleVal.toFixed(2);
    if ('boolVal' in cell) return cell.boolVal ? 'true' : 'false';
    if ('bytesVal' in cell) return '<bytes>';
    return '';
  };

  return (
    <div
      className="rounded-lg overflow-hidden"
      style={{ border: '1px solid var(--border-primary)' }}
    >
      <div className="max-h-64 overflow-auto">
        <div className="flex sticky top-0 z-10" style={{ background: 'var(--bg-tertiary)' }}>
          {columns.map((col) => (
            <div
              key={col}
              className="px-3 py-2 text-xs font-mono font-bold flex-1 min-w-[120px]"
              style={{ color: 'var(--text-primary)' }}
            >
              {col}
            </div>
          ))}
        </div>
        {rows.map((row, i) => (
          <div key={i} className="flex" style={{ borderBottom: '1px solid var(--border-primary)' }}>
            {columns.map((col) => (
              <div
                key={col}
                className="px-3 py-1.5 text-xs font-mono flex-1 min-w-[120px] truncate"
                style={{ color: 'var(--text-secondary)' }}
                title={cellText(row.cells[col])}
              >
                {cellText(row.cells[col])}
              </div>
            ))}
          </div>
        ))}
      </div>
      <div
        className="flex items-center justify-between px-3 py-2 text-xs font-mono"
        style={{ background: 'var(--bg-tertiary)' }}
      >
        <span style={{ color: 'var(--text-tertiary)' }}>
          {rows.length} of {totalFetched} rows
        </span>
        {hasMore && (
          <button
            onClick={onLoadMore}
            className="px-3 py-1 rounded"
            style={{ background: 'var(--accent-subtle)', color: 'var(--accent-primary)' }}
          >
            Load more
          </button>
        )}
      </div>
    </div>
  );
}
