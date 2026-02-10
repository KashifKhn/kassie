import { useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { List } from 'react-window';
import { ChevronLeft, ChevronRight, Loader2 } from 'lucide-react';
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

  const { data: rowsData, isLoading } = useQuery({
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
      <div className="h-full flex items-center justify-center">
        <div className="flex flex-col items-center gap-2">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
          <p className="text-sm text-muted-foreground">Loading data...</p>
        </div>
      </div>
    );
  }

  if (!schemaData || !rowsData) {
    return (
      <div className="h-full flex items-center justify-center text-muted-foreground">
        No data available
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

  return (
    <div className="h-full flex flex-col overflow-hidden bg-background">
      <div className="flex-shrink-0 border-b border-border bg-muted">
        <div className="flex">
          {columns.map((column) => (
            <div
              key={column.name}
              className="px-4 py-2 text-sm font-semibold border-r border-border flex-1 min-w-[150px]"
            >
              <div className="flex items-center gap-2">
                <span>{column.name}</span>
                <span className="text-xs text-muted-foreground">
                  {column.type}
                </span>
                {column.isPartitionKey && (
                  <span className="text-xs bg-primary text-primary-foreground px-1 rounded">
                    PK
                  </span>
                )}
                {column.isClusteringKey && (
                  <span className="text-xs bg-secondary text-secondary-foreground px-1 rounded">
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

      <div className="flex-shrink-0 border-t border-border px-4 py-2 bg-muted">
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">
            Showing {displayRows.length} {displayRows.length === 1 ? 'row' : 'rows'}
          </span>
          <div className="flex items-center gap-2">
            {allRows.length > pageSize && (
              <button
                onClick={handlePrevPage}
                className="flex items-center gap-1 px-3 py-1 text-sm font-medium bg-primary text-primary-foreground rounded hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ChevronLeft className="h-4 w-4" />
                Previous
              </button>
            )}
            {currentHasMore && (
              <button
                onClick={handleNextPage}
                disabled={nextPageMutation.isPending}
                className="flex items-center gap-1 px-3 py-1 text-sm font-medium bg-primary text-primary-foreground rounded hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
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
      style={style}
      className="flex border-b border-border hover:bg-accent cursor-pointer transition-colors"
      onClick={() => onRowSelect?.(row)}
    >
      {columns.map((column) => (
        <div
          key={column.name}
          className="px-4 py-2 text-sm border-r border-border flex-1 min-w-[150px] truncate"
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
