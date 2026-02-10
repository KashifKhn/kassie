import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ChevronRight, ChevronDown, Database, Table, Loader2 } from 'lucide-react';
import { schemaApi, queryKeys } from '@/api/queries';
import { useUiStore } from '@/stores/uiStore';

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
      <div 
        className="h-full flex items-center justify-center"
        style={{ background: 'var(--bg-primary)' }}
      >
        <div className="flex flex-col items-center gap-3">
          <Loader2 
            className="h-6 w-6 animate-spin" 
            style={{ color: 'var(--accent-primary)' }}
          />
          <p 
            className="text-sm font-mono"
            style={{ color: 'var(--text-secondary)' }}
          >
            Loading schema...
          </p>
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
        className="px-5 py-4"
        style={{ 
          borderBottom: '1px solid var(--border-primary)',
          background: 'var(--bg-secondary)'
        }}
      >
        <h2 
          className="text-sm font-mono font-bold tracking-wider uppercase"
          style={{ color: 'var(--text-primary)' }}
        >
          Schema Explorer
        </h2>
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

  const isSelected = selectedKeyspace === keyspace;

  return (
    <div className="animate-fade-in">
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-3 px-5 py-3 text-sm font-mono transition-all"
        style={{
          background: isSelected ? 'var(--accent-subtle)' : 'transparent',
          color: isSelected ? 'var(--accent-primary)' : 'var(--text-primary)',
          borderLeft: isSelected ? '3px solid var(--accent-primary)' : '3px solid transparent',
        }}
        onMouseEnter={(e) => {
          if (!isSelected) {
            e.currentTarget.style.background = 'var(--bg-secondary)';
          }
        }}
        onMouseLeave={(e) => {
          if (!isSelected) {
            e.currentTarget.style.background = 'transparent';
          }
        }}
      >
        {isExpanded ? (
          <ChevronDown 
            className="w-4 h-4 flex-shrink-0 transition-transform" 
            style={{ color: 'var(--text-tertiary)' }}
          />
        ) : (
          <ChevronRight 
            className="w-4 h-4 flex-shrink-0 transition-transform" 
            style={{ color: 'var(--text-tertiary)' }}
          />
        )}
        <Database 
          className="w-5 h-5 flex-shrink-0" 
          style={{ color: 'var(--accent-primary)' }}
        />
        <span className="flex-1 text-left font-semibold break-words">{keyspace}</span>
      </button>

      {isExpanded && tablesData && (
        <div className="animate-slide-down" style={{ background: 'var(--bg-secondary)' }}>
          {tablesData.tables.map((table, index) => {
            const isTableSelected = selectedKeyspace === keyspace && selectedTable === table.name;
            return (
              <button
                key={table.name}
                onClick={() => onTableSelect?.(keyspace, table.name)}
                className="w-full flex items-center gap-3 px-5 py-2.5 text-sm font-mono transition-all animate-fade-in"
                style={{
                  background: isTableSelected ? 'var(--accent-subtle)' : 'transparent',
                  color: isTableSelected ? 'var(--accent-primary)' : 'var(--text-secondary)',
                  borderLeft: isTableSelected ? '3px solid var(--accent-primary)' : '3px solid transparent',
                  paddingLeft: '2.5rem',
                  animationDelay: `${index * 50}ms`,
                }}
                onMouseEnter={(e) => {
                  if (!isTableSelected) {
                    e.currentTarget.style.background = 'var(--bg-tertiary)';
                    e.currentTarget.style.color = 'var(--text-primary)';
                  }
                }}
                onMouseLeave={(e) => {
                  if (!isTableSelected) {
                    e.currentTarget.style.background = 'transparent';
                    e.currentTarget.style.color = 'var(--text-secondary)';
                  }
                }}
              >
                <Table 
                  className="w-4 h-4 flex-shrink-0" 
                  style={{ color: 'var(--text-tertiary)' }}
                />
                <span className="flex-1 text-left break-words">{table.name}</span>
                {table.estimatedRows > 0 && (
                  <span 
                    className="text-xs font-mono flex-shrink-0 ml-2"
                    style={{ color: 'var(--text-tertiary)' }}
                  >
                    {formatNumber(table.estimatedRows)}
                  </span>
                )}
              </button>
            );
          })}
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
