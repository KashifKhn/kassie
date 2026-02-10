import { useQuery } from '@tanstack/react-query';
import { List } from 'react-window';
import { dataApi, queryKeys, schemaApi } from '@/api/queries';
import { useUiStore } from '@/stores/uiStore';
import type { Row, CellValue } from '@/api/types';

interface DataGridProps {
  keyspace: string;
  table: string;
  onRowSelect?: (row: Row) => void;
}

export function DataGrid({ keyspace, table, onRowSelect }: DataGridProps) {
  const { pageSize } = useUiStore();

  const { data: schemaData } = useQuery({
    queryKey: queryKeys.schema.tableSchema(keyspace, table),
    queryFn: () => schemaApi.getTableSchema(keyspace, table),
  });

  const { data: rowsData, isLoading } = useQuery({
    queryKey: queryKeys.data.rows(keyspace, table, pageSize),
    queryFn: () => dataApi.queryRows({ keyspace, table, pageSize }),
  });

  if (isLoading) {
    return (
      <div className="h-full flex items-center justify-center text-muted-foreground">
        Loading data...
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
  const rows = rowsData.rows;

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
          rowCount={rows.length}
          rowHeight={40}
          rowComponent={RowRenderer}
          rowProps={{ rows, columns, onRowSelect }}
        />
      </div>

      <div className="flex-shrink-0 border-t border-border px-4 py-2 bg-muted">
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <span>
            Showing {rows.length} {rows.length === 1 ? 'row' : 'rows'}
          </span>
          {rowsData.hasMore && (
            <span className="text-primary">More data available</span>
          )}
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
