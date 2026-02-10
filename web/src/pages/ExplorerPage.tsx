import { useState } from 'react';
import { Layout } from '@/components/layout/Layout';
import { Header } from '@/components/header/Header';
import { Sidebar } from '@/components/sidebar/Sidebar';
import { FilterBar } from '@/components/filterbar/FilterBar';
import { DataGrid } from '@/components/datagrid/DataGrid';
import { Inspector } from '@/components/inspector/Inspector';
import { useUiStore } from '@/stores/uiStore';
import type { Row } from '@/api/types';

export function ExplorerPage() {
  const { selectedKeyspace, selectedTable } = useUiStore();
  const [selectedRow, setSelectedRow] = useState<Row | null>(null);

  const handleRowSelect = (row: Row) => {
    setSelectedRow(row);
  };

  const handleFilter = (_filter: string) => {
    // TODO: Implement filtering in Phase 5
  };

  const handleClearFilter = () => {
    // TODO: Implement filter clearing in Phase 5
  };

  return (
    <Layout
      header={<Header />}
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
