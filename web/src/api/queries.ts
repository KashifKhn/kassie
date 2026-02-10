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
};
