import { useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { List } from 'react-window';
import { ChevronLeft, ChevronRight, Loader2, AlertCircle, Database } from 'lucide-react';
import { dataApi, queryKeys, schemaApi } from '@/api/queries';
import { useUiStore } from '@/stores/uiStore';
import type { Row, CellValue } from '@/api/types';

interface DataGridProps {
  keyspace: string;
  table: string;
  whereClause?: string;
  onRowSelect?: (row: Row) => void;
}

export function DataGrid({
  keyspace,
  table,
  whereClause,
  onRowSelect,
}: DataGridProps) {
  const { pageSize } = useUiStore();
  const [cursorId, setCursorId] = useState<string | null>(null);
  const [allRows, setAllRows] = useState<Row[]>([]);
  const [hasMore, setHasMore] = useState(false);

  const { data: schemaData } = useQuery({
    queryKey: queryKeys.schema.tableSchema(keyspace, table),
    queryFn: () => schemaApi.getTableSchema(keyspace, table),
  });

  const isFiltered = Boolean(whereClause?.trim());

  const { data: rowsData, isLoading, error } = useQuery({
    queryKey: isFiltered
      ? queryKeys.data.filteredRows(keyspace, table, whereClause || '', pageSize)
      : queryKeys.data.rows(keyspace, table, pageSize),
    queryFn: () =>
      isFiltered
        ? dataApi.filterRows({
            keyspace,
            table,
            whereClause: whereClause || '',
            pageSize,
          })
        : dataApi.queryRows({ keyspace, table, pageSize }),
    enabled: Boolean(keyspace && table),
  });

  const nextPageMutation = useMutation({
    mutationFn: dataApi.getNextPage,
    onSuccess: (data) => {
      setAllRows((prev) => [...prev, ...data.rows]);
      setCursorId(data.cursorId);
      setHasMore(data.hasMore);
    },
  });

  const handleNextPage = () => {
    if (cursorId && hasMore) {
      nextPageMutation.mutate({ cursorId });
    }
  };

  const handlePrevPage = () => {
    if (allRows.length > pageSize) {
      setAllRows((prev) => prev.slice(0, -pageSize));
      setHasMore(true);
    }
  };

  if (isLoading) {
    return (
      <div 
        className="h-full flex items-center justify-center noise-bg"
        style={{ background: 'var(--bg-primary)' }}
      >
        <div className="flex flex-col items-center gap-3 animate-fade-in">
          <Loader2 
            className="h-10 w-10 animate-spin" 
            style={{ 
              color: 'var(--accent-primary)',
              filter: 'drop-shadow(0 0 10px var(--accent-primary))',
            }}
          />
          <p 
            className="text-sm font-mono"
            style={{ color: 'var(--text-secondary)' }}
          >
            Loading data...
          </p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div 
        className="h-full flex items-center justify-center"
        style={{ background: 'var(--bg-primary)' }}
      >
        <div 
          className="flex flex-col items-center gap-4 max-w-md text-center p-8 rounded-xl glass animate-scale-in"
          style={{
            border: '1px solid var(--border-primary)',
            boxShadow: 'var(--shadow-lg)',
          }}
        >
          <AlertCircle 
            className="h-12 w-12" 
            style={{ 
              color: 'var(--error)',
              filter: 'drop-shadow(0 0 20px var(--error))',
            }}
          />
          <div>
            <h3 
              className="text-lg font-mono font-bold"
              style={{ color: 'var(--text-primary)' }}
            >
              Failed to load data
            </h3>
            <p 
              className="mt-3 text-sm font-sans"
              style={{ color: 'var(--text-secondary)' }}
            >
              {error instanceof Error ? error.message : 'An error occurred while fetching data'}
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (!schemaData || !rowsData) {
    return (
      <div 
        className="h-full flex items-center justify-center"
        style={{ background: 'var(--bg-primary)' }}
      >
        <div className="flex flex-col items-center gap-3 animate-fade-in">
          <Database 
            className="h-12 w-12" 
            style={{ color: 'var(--text-tertiary)' }}
          />
          <p 
            className="text-sm font-mono"
            style={{ color: 'var(--text-secondary)' }}
          >
            No data available
          </p>
        </div>
      </div>
    );
  }

  const columns = schemaData.schema.columns;
  const displayRows = allRows.length > 0 ? allRows : rowsData.rows;
  const currentHasMore = allRows.length > 0 ? hasMore : rowsData.hasMore;
  const currentCursorId = cursorId || rowsData.cursorId;

  if (cursorId !== currentCursorId) {
    setCursorId(currentCursorId);
    setHasMore(currentHasMore);
    if (allRows.length === 0) {
      setAllRows(rowsData.rows);
    }
  }

  if (displayRows.length === 0) {
    return (
      <div 
        className="h-full flex flex-col overflow-hidden"
        style={{ background: 'var(--bg-primary)' }}
      >
        <div 
          className="flex-shrink-0"
          style={{ 
            borderBottom: '1px solid var(--border-primary)',
            background: 'var(--bg-secondary)',
          }}
        >
          <div className="flex">
            {columns.map((column) => (
              <div
                key={column.name}
                className="px-4 py-3 text-sm font-mono font-bold border-r flex-1 min-w-[150px]"
                style={{ 
                  color: 'var(--text-primary)',
                  borderRight: '1px solid var(--border-primary)',
                }}
              >
                <div className="flex items-center gap-2">
                  <span>{column.name}</span>
                  <span 
                    className="text-xs"
                    style={{ color: 'var(--text-tertiary)' }}
                  >
                    {column.type}
                  </span>
                  {column.isPartitionKey && (
                    <span 
                      className="text-xs px-1.5 py-0.5 rounded font-bold"
                      style={{ 
                        background: 'var(--accent-primary)',
                        color: 'var(--text-inverse)',
                      }}
                    >
                      PK
                    </span>
                  )}
                  {column.isClusteringKey && (
                    <span 
                      className="text-xs px-1.5 py-0.5 rounded font-bold"
                      style={{ 
                        background: 'var(--accent-subtle)',
                        color: 'var(--accent-primary)',
                      }}
                    >
                      CK
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
        <div 
          className="flex-1 flex items-center justify-center noise-bg"
          style={{ background: 'var(--bg-primary)' }}
        >
          <div className="flex flex-col items-center gap-3 animate-fade-in">
            <Database 
              className="h-12 w-12" 
              style={{ color: 'var(--text-tertiary)' }}
            />
            <p 
              className="text-sm font-mono"
              style={{ color: 'var(--text-secondary)' }}
            >
              {isFiltered ? 'No rows match the filter' : 'No data in this table'}
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div 
      className="h-full flex flex-col overflow-hidden"
      style={{ background: 'var(--bg-primary)' }}
    >
      <div 
        className="flex-shrink-0"
        style={{ 
          borderBottom: '1px solid var(--border-primary)',
          background: 'var(--bg-secondary)',
        }}
      >
        <div className="flex">
          {columns.map((column) => (
            <div
              key={column.name}
              className="px-4 py-3 text-sm font-mono font-bold border-r flex-1 min-w-[150px]"
              style={{ 
                color: 'var(--text-primary)',
                borderRight: '1px solid var(--border-primary)',
              }}
            >
              <div className="flex items-center gap-2">
                <span>{column.name}</span>
                <span 
                  className="text-xs"
                  style={{ color: 'var(--text-tertiary)' }}
                >
                  {column.type}
                </span>
                {column.isPartitionKey && (
                  <span 
                    className="text-xs px-1.5 py-0.5 rounded font-bold"
                    style={{ 
                      background: 'var(--accent-primary)',
                      color: 'var(--text-inverse)',
                    }}
                  >
                    PK
                  </span>
                )}
                {column.isClusteringKey && (
                  <span 
                    className="text-xs px-1.5 py-0.5 rounded font-bold"
                    style={{ 
                      background: 'var(--accent-subtle)',
                      color: 'var(--accent-primary)',
                    }}
                  >
                    CK
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="flex-1">
        <List<RowData>
          defaultHeight={600}
          rowCount={displayRows.length}
          rowHeight={40}
          rowComponent={RowRenderer}
          rowProps={{ rows: displayRows, columns, onRowSelect }}
        />
      </div>

      <div 
        className="flex-shrink-0 px-4 py-3"
        style={{ 
          borderTop: '1px solid var(--border-primary)',
          background: 'var(--bg-secondary)',
        }}
      >
        <div className="flex items-center justify-between text-sm font-mono">
          <span style={{ color: 'var(--text-secondary)' }}>
            Showing <span style={{ color: 'var(--accent-primary)' }}>{displayRows.length}</span> {displayRows.length === 1 ? 'row' : 'rows'}
          </span>
          <div className="flex items-center gap-2">
            {allRows.length > pageSize && (
              <button
                onClick={handlePrevPage}
                className="flex items-center gap-1 px-4 py-2 text-sm font-mono font-medium rounded transition-all"
                style={{
                  background: 'var(--accent-primary)',
                  color: 'var(--text-inverse)',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'var(--accent-hover)';
                  e.currentTarget.style.transform = 'translateY(-1px)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'var(--accent-primary)';
                  e.currentTarget.style.transform = 'translateY(0)';
                }}
              >
                <ChevronLeft className="h-4 w-4" />
                Previous
              </button>
            )}
            {currentHasMore && (
              <button
                onClick={handleNextPage}
                disabled={nextPageMutation.isPending}
                className="flex items-center gap-1 px-4 py-2 text-sm font-mono font-medium rounded transition-all"
                style={{
                  background: nextPageMutation.isPending ? 'var(--bg-tertiary)' : 'var(--accent-primary)',
                  color: nextPageMutation.isPending ? 'var(--text-tertiary)' : 'var(--text-inverse)',
                  cursor: nextPageMutation.isPending ? 'not-allowed' : 'pointer',
                  opacity: nextPageMutation.isPending ? '0.5' : '1',
                }}
                onMouseEnter={(e) => {
                  if (!nextPageMutation.isPending) {
                    e.currentTarget.style.background = 'var(--accent-hover)';
                    e.currentTarget.style.transform = 'translateY(-1px)';
                  }
                }}
                onMouseLeave={(e) => {
                  if (!nextPageMutation.isPending) {
                    e.currentTarget.style.background = 'var(--accent-primary)';
                    e.currentTarget.style.transform = 'translateY(0)';
                  }
                }}
              >
                Next
                <ChevronRight className="h-4 w-4" />
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

interface RowData {
  rows: Row[];
  columns: Array<{ name: string }>;
  onRowSelect?: (row: Row) => void;
}

function RowRenderer({
  index,
  style,
  rows,
  columns,
  onRowSelect,
}: {
  index: number;
  style: React.CSSProperties;
} & RowData) {
  const row = rows[index];

  if (!row) return null;

  return (
    <div
      style={{
        ...style,
        borderBottom: '1px solid var(--border-primary)',
        background: 'transparent',
        cursor: 'pointer',
      }}
      className="flex transition-all"
      onClick={() => onRowSelect?.(row)}
      onMouseEnter={(e) => {
        e.currentTarget.style.background = 'var(--accent-subtle)';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.background = 'transparent';
      }}
    >
      {columns.map((column) => (
        <div
          key={column.name}
          className="px-4 py-2 text-sm font-mono border-r flex-1 min-w-[150px] truncate"
          style={{
            color: 'var(--text-primary)',
            borderRight: '1px solid var(--border-primary)',
          }}
        >
          {formatCellValue(row.cells[column.name])}
        </div>
      ))}
    </div>
  );
}

function formatCellValue(value: CellValue | undefined): string {
  if (!value || value.isNull) {
    return 'NULL';
  }

  if ('stringVal' in value) return value.stringVal;
  if ('intVal' in value) return value.intVal.toString();
  if ('doubleVal' in value) return value.doubleVal.toFixed(2);
  if ('boolVal' in value) return value.boolVal ? 'true' : 'false';
  if ('bytesVal' in value) return '<bytes>';

  return '';
}
