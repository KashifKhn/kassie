import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type Theme = 'light' | 'dark' | 'system';

interface UiState {
  theme: Theme;
  sidebarCollapsed: boolean;
  inspectorCollapsed: boolean;
  filterBarVisible: boolean;
  pageSize: number;
  selectedKeyspace: string | null;
  selectedTable: string | null;
}

interface UiActions {
  setTheme: (theme: Theme) => void;
  toggleSidebar: () => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  toggleInspector: () => void;
  setInspectorCollapsed: (collapsed: boolean) => void;
  setFilterBarVisible: (visible: boolean) => void;
  setPageSize: (size: number) => void;
  setSelectedKeyspace: (keyspace: string | null) => void;
  setSelectedTable: (table: string | null) => void;
}

const initialState: UiState = {
  theme: 'system',
  sidebarCollapsed: false,
  inspectorCollapsed: false,
  filterBarVisible: true,
  pageSize: 100,
  selectedKeyspace: null,
  selectedTable: null,
};

export const useUiStore = create<UiState & UiActions>()(
  persist(
    (set) => ({
      ...initialState,

      setTheme: (theme) => {
        set({ theme });
      },

      toggleSidebar: () => {
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed }));
      },

      setSidebarCollapsed: (collapsed) => {
        set({ sidebarCollapsed: collapsed });
      },

      toggleInspector: () => {
        set((state) => ({ inspectorCollapsed: !state.inspectorCollapsed }));
      },

      setInspectorCollapsed: (collapsed) => {
        set({ inspectorCollapsed: collapsed });
      },

      setFilterBarVisible: (visible) => {
        set({ filterBarVisible: visible });
      },

      setPageSize: (size) => {
        if (size < 10 || size > 1000) return;
        set({ pageSize: size });
      },

      setSelectedKeyspace: (keyspace) => {
        set({ selectedKeyspace: keyspace, selectedTable: null });
      },

      setSelectedTable: (table) => {
        set({ selectedTable: table });
      },
    }),
    {
      name: 'ui-storage',
    }
  )
);
