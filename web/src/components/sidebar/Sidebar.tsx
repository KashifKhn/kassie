import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ChevronRight, ChevronDown, Database, Table, Loader2 } from 'lucide-react';
import { schemaApi, queryKeys } from '@/api/queries';
import { useUiStore } from '@/stores/uiStore';
import { cn } from '@/lib/utils';

export function Sidebar() {
  const { selectedKeyspace, selectedTable, setSelectedKeyspace, setSelectedTable } = useUiStore();
  const [expandedKeyspaces, setExpandedKeyspaces] = useState<Set<string>>(
    new Set()
  );

  const { data: keyspacesData, isLoading } = useQuery({
    queryKey: queryKeys.schema.keyspaces(),
    queryFn: schemaApi.listKeyspaces,
  });

  const toggleKeyspace = (keyspace: string) => {
    setExpandedKeyspaces((prev) => {
      const next = new Set(prev);
      if (next.has(keyspace)) {
        next.delete(keyspace);
      } else {
        next.add(keyspace);
      }
      return next;
    });
  };

  const handleTableSelect = (keyspace: string, table: string) => {
    setSelectedKeyspace(keyspace);
    setSelectedTable(table);
  };

  if (isLoading) {
    return (
      <div className="h-full flex items-center justify-center">
        <div className="flex flex-col items-center gap-2">
          <Loader2 className="h-6 w-6 animate-spin text-primary" />
          <p className="text-sm text-muted-foreground">Loading schema...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="p-4 border-b border-border">
        <h2 className="text-sm font-semibold text-foreground">Schema</h2>
      </div>

      <div className="flex-1 overflow-y-auto">
        {keyspacesData?.keyspaces.map((keyspace) => (
          <KeyspaceNode
            key={keyspace.name}
            keyspace={keyspace.name}
            isExpanded={expandedKeyspaces.has(keyspace.name)}
            onToggle={() => toggleKeyspace(keyspace.name)}
            onTableSelect={handleTableSelect}
            selectedKeyspace={selectedKeyspace || undefined}
            selectedTable={selectedTable || undefined}
          />
        ))}
      </div>
    </div>
  );
}

interface KeyspaceNodeProps {
  keyspace: string;
  isExpanded: boolean;
  onToggle: () => void;
  onTableSelect?: (keyspace: string, table: string) => void;
  selectedKeyspace?: string;
  selectedTable?: string;
}

function KeyspaceNode({
  keyspace,
  isExpanded,
  onToggle,
  onTableSelect,
  selectedKeyspace,
  selectedTable,
}: KeyspaceNodeProps) {
  const { data: tablesData } = useQuery({
    queryKey: queryKeys.schema.tables(keyspace),
    queryFn: () => schemaApi.listTables(keyspace),
    enabled: isExpanded,
  });

  return (
    <div>
      <button
        onClick={onToggle}
        className={cn(
          'w-full flex items-center gap-2 px-4 py-2 text-sm hover:bg-accent transition-colors',
          selectedKeyspace === keyspace && 'bg-accent'
        )}
      >
        {isExpanded ? (
          <ChevronDown className="w-4 h-4 text-muted-foreground" />
        ) : (
          <ChevronRight className="w-4 h-4 text-muted-foreground" />
        )}
        <Database className="w-4 h-4 text-primary" />
        <span className="flex-1 text-left font-medium">{keyspace}</span>
      </button>

      {isExpanded && tablesData && (
        <div className="ml-6">
          {tablesData.tables.map((table) => (
            <button
              key={table.name}
              onClick={() => onTableSelect?.(keyspace, table.name)}
              className={cn(
                'w-full flex items-center gap-2 px-4 py-2 text-sm hover:bg-accent transition-colors',
                selectedKeyspace === keyspace &&
                  selectedTable === table.name &&
                  'bg-accent text-primary'
              )}
            >
              <Table className="w-4 h-4 text-muted-foreground" />
              <span className="flex-1 text-left">{table.name}</span>
              <span className="text-xs text-muted-foreground">
                {table.estimatedRows > 0 && `~${formatNumber(table.estimatedRows)}`}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

function formatNumber(num: number): string {
  if (num >= 1_000_000) {
    return `${(num / 1_000_000).toFixed(1)}M`;
  }
  if (num >= 1_000) {
    return `${(num / 1_000).toFixed(1)}K`;
  }
  return num.toString();
}
