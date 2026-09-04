import { QueryClient } from '@tanstack/react-query';
import { apiClient, handleApiError } from './client';
import { LoginResponseSchema, RefreshResponseSchema, GetProfilesResponseSchema } from './schemas';
import type {
  GetProfilesResponse,
  LoginRequest,
  LoginResponse,
  RefreshRequest,
  RefreshResponse,
  ListKeyspacesResponse,
  ListTablesResponse,
  GetTableSchemaResponse,
  QueryRowsRequest,
  QueryRowsResponse,
  GetNextPageRequest,
  GetNextPageResponse,
  FilterRowsRequest,
  FilterRowsResponse,
  TableStats,
  ExecuteQueryRequest,
  ExecuteQueryResponse,
  QueryHistoryEntry,
  SavedQuery,
  SlowQuery,
} from './types';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      gcTime: 10 * 60 * 1000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
});

export const queryKeys = {
  auth: {
    profiles: () => ['profiles'] as const,
  },
  schema: {
    keyspaces: () => ['keyspaces'] as const,
    tables: (keyspace: string) => ['tables', keyspace] as const,
    tableSchema: (keyspace: string, table: string) =>
      ['tableSchema', keyspace, table] as const,
  },
  history: {
    queries: () => ['history', 'queries'] as const,
    saved: () => ['history', 'saved'] as const,
    slow: () => ['history', 'slow'] as const,
  },
  stats: (keyspace: string, table: string) => ['stats', keyspace, table] as const,
  data: {
    rows: (keyspace: string, table: string, pageSize: number) =>
      ['rows', keyspace, table, pageSize] as const,
    filteredRows: (
      keyspace: string,
      table: string,
      whereClause: string,
      pageSize: number
    ) => ['filteredRows', keyspace, table, whereClause, pageSize] as const,
  },
};

export const sessionApi = {
  login: async (request: LoginRequest): Promise<LoginResponse> => {
    try {
      const response = await apiClient.post('/session/login', request);
      return LoginResponseSchema.parse(response.data);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  refresh: async (request: RefreshRequest): Promise<RefreshResponse> => {
    try {
      const response = await apiClient.post('/session/refresh', request);
      return RefreshResponseSchema.parse(response.data);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  logout: async (): Promise<void> => {
    try {
      await apiClient.post('/session/logout', {});
    } catch (error) {
      throw handleApiError(error);
    }
  },

  getProfiles: async (): Promise<GetProfilesResponse> => {
    try {
      const response = await apiClient.get('/profiles');
      return GetProfilesResponseSchema.parse(response.data);
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export const schemaApi = {
  listKeyspaces: async (): Promise<ListKeyspacesResponse> => {
    try {
      const response = await apiClient.get<ListKeyspacesResponse>(
        '/schema/keyspaces'
      );
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  listTables: async (keyspace: string): Promise<ListTablesResponse> => {
    try {
      const response = await apiClient.get<ListTablesResponse>(
        `/schema/keyspaces/${keyspace}/tables`
      );
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  getTableSchema: async (
    keyspace: string,
    table: string
  ): Promise<GetTableSchemaResponse> => {
    try {
      const response = await apiClient.get<GetTableSchemaResponse>(
        `/schema/keyspaces/${keyspace}/tables/${table}`
      );
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export const dataApi = {
  queryRows: async (request: QueryRowsRequest): Promise<QueryRowsResponse> => {
    try {
      const response = await apiClient.post<QueryRowsResponse>(
        '/data/query',
        request
      );
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  getNextPage: async (
    request: GetNextPageRequest
  ): Promise<GetNextPageResponse> => {
    try {
      const response = await apiClient.post<GetNextPageResponse>(
        '/data/next',
        request
      );
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  filterRows: async (
    request: FilterRowsRequest
  ): Promise<FilterRowsResponse> => {
    try {
      const response = await apiClient.post<FilterRowsResponse>(
        '/data/filter',
        request
      );
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  executeQuery: async (
    request: ExecuteQueryRequest
  ): Promise<ExecuteQueryResponse> => {
    try {
      const response = await apiClient.post<ExecuteQueryResponse>(
        '/data/cql',
        request
      );
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export const statsApi = {
  getTableStats: async (keyspace: string, table: string): Promise<TableStats> => {
    try {
      const response = await apiClient.post<{ stats: TableStats }>('/schema/stats', {
        keyspace,
        table,
      });
      return response.data.stats;
    } catch (error) {
      throw handleApiError(error);
    }
  },
};

export const historyApi = {
  listHistory: async (limit: number): Promise<QueryHistoryEntry[]> => {
    try {
      const response = await apiClient.get<{ entries: QueryHistoryEntry[] }>(
        '/history/queries',
        { params: { limit } }
      );
      return response.data.entries ?? [];
    } catch (error) {
      throw handleApiError(error);
    }
  },

  clearHistory: async (): Promise<void> => {
    try {
      await apiClient.post('/history/queries/clear');
    } catch (error) {
      throw handleApiError(error);
    }
  },

  saveQuery: async (name: string, cql: string): Promise<SavedQuery> => {
    try {
      const response = await apiClient.post<{ query: SavedQuery }>(
        '/history/saved',
        { name, cql }
      );
      return response.data.query;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  listSavedQueries: async (): Promise<SavedQuery[]> => {
    try {
      const response = await apiClient.get<{ queries: SavedQuery[] }>(
        '/history/saved'
      );
      return response.data.queries ?? [];
    } catch (error) {
      throw handleApiError(error);
    }
  },

  deleteSavedQuery: async (name: string): Promise<void> => {
    try {
      await apiClient.delete(`/history/saved/${encodeURIComponent(name)}`);
    } catch (error) {
      throw handleApiError(error);
    }
  },

  getSlowQueries: async (limit: number): Promise<SlowQuery[]> => {
    try {
      const response = await apiClient.get<{ queries: SlowQuery[] }>(
        '/history/slow',
        { params: { limit } }
      );
      return response.data.queries ?? [];
    } catch (error) {
      throw handleApiError(error);
    }
  },
};
