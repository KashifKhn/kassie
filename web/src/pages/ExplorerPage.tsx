import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { Database, ArrowRight } from 'lucide-react';
import { Layout } from '@/components/layout/Layout';
import { Header } from '@/components/header/Header';
import { Sidebar } from '@/components/sidebar/Sidebar';
import { FilterBar } from '@/components/filterbar/FilterBar';
import { DataGrid } from '@/components/datagrid/DataGrid';
import { Inspector } from '@/components/inspector/Inspector';
import { useUiStore } from '@/stores/uiStore';
import { useAuthStore } from '@/stores/authStore';
import { useToastStore } from '@/stores/toastStore';
import { sessionApi } from '@/api/queries';
import type { Row } from '@/api/types';

export function ExplorerPage() {
  const navigate = useNavigate();
  const { selectedKeyspace, selectedTable } = useUiStore();
  const { clearAuth } = useAuthStore();
  const { success } = useToastStore();
  const [selectedRow, setSelectedRow] = useState<Row | null>(null);
  const [whereClause, setWhereClause] = useState<string>('');

  const logoutMutation = useMutation({
    mutationFn: sessionApi.logout,
    onSuccess: () => {
      clearAuth();
      success('Logged out successfully');
      navigate('/login');
    },
  });

  const handleLogout = () => {
    logoutMutation.mutate();
  };

  const handleRowSelect = (row: Row) => {
    setSelectedRow(row);
  };

  const handleFilter = (filter: string) => {
    setWhereClause(filter);
  };

  const handleClearFilter = () => {
    setWhereClause('');
  };

  return (
    <Layout
      header={<Header onLogout={handleLogout} />}
      sidebar={<Sidebar />}
      main={
        <div 
          className="flex h-full flex-col"
          style={{ background: 'var(--bg-primary)' }}
        >
          {selectedKeyspace && selectedTable ? (
            <>
              <FilterBar onFilter={handleFilter} onClear={handleClearFilter} />
              <div className="flex-1 overflow-hidden">
              <DataGrid
                keyspace={selectedKeyspace}
                table={selectedTable}
                whereClause={whereClause}
                onRowSelect={handleRowSelect}
              />
              </div>
            </>
          ) : (
            <div className="flex h-full items-center justify-center noise-bg">
              <div 
                className="text-center p-12 rounded-xl glass animate-scale-in"
                style={{
                  maxWidth: '500px',
                  border: '1px solid var(--border-primary)',
                  boxShadow: 'var(--shadow-lg)',
                }}
              >
                <Database 
                  className="h-16 w-16 mx-auto mb-6 animate-pulse"
                  style={{ 
                    color: 'var(--accent-primary)',
                    filter: 'drop-shadow(0 0 20px var(--accent-primary))',
                  }}
                />
                <h2 
                  className="text-2xl font-mono font-bold mb-3"
                  style={{ color: 'var(--text-primary)' }}
                >
                  No Table Selected
                </h2>
                <p 
                  className="text-sm font-sans mb-6 leading-relaxed"
                  style={{ color: 'var(--text-secondary)' }}
                >
                  Select a keyspace and table from the sidebar to begin exploring your data
                </p>
                <div 
                  className="flex items-center justify-center gap-2 text-xs font-mono"
                  style={{ color: 'var(--text-tertiary)' }}
                >
                  <Database className="w-3 h-3" />
                  <ArrowRight className="w-3 h-3" />
                  <span>Choose keyspace</span>
                  <ArrowRight className="w-3 h-3" />
                  <span>Select table</span>
                </div>
              </div>
            </div>
          )}
        </div>
      }
      inspector={<Inspector row={selectedRow} />}
    />
  );
}
