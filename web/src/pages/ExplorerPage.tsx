import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
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
        <div className="flex h-full flex-col">
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
            <div className="flex h-full items-center justify-center">
              <div className="text-center text-gray-500 dark:text-gray-400">
                <p className="text-lg font-medium">No table selected</p>
                <p className="mt-2 text-sm">
                  Select a table from the sidebar to view data
                </p>
              </div>
            </div>
          )}
        </div>
      }
      inspector={<Inspector row={selectedRow} />}
    />
  );
}
