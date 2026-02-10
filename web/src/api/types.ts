export interface ProfileInfo {
  name: string;
  hosts: string[];
  port: number;
  keyspace: string;
  sslEnabled: boolean;
}

export interface LoginRequest {
  profile: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
  profile: ProfileInfo;
}

export interface RefreshRequest {
  refreshToken: string;
}

export interface RefreshResponse {
  accessToken: string;
  expiresAt: number;
}

export interface LogoutRequest {}

export interface LogoutResponse {}

export interface GetProfilesRequest {}

export interface GetProfilesResponse {
  profiles: ProfileInfo[];
}

export interface Keyspace {
  name: string;
  replicationStrategy: string;
  replication: Record<string, string>;
}

export interface Table {
  name: string;
  keyspace: string;
  estimatedRows: number;
}

export interface Column {
  name: string;
  type: string;
  isPartitionKey: boolean;
  isClusteringKey: boolean;
  position: number;
}

export interface TableSchema {
  keyspace: string;
  table: string;
  columns: Column[];
  partitionKeys: string[];
  clusteringKeys: string[];
}

export interface ListKeyspacesRequest {}

export interface ListKeyspacesResponse {
  keyspaces: Keyspace[];
}

export interface ListTablesRequest {
  keyspace: string;
}

export interface ListTablesResponse {
  tables: Table[];
}

export interface GetTableSchemaRequest {
  keyspace: string;
  table: string;
}

export interface GetTableSchemaResponse {
  schema: TableSchema;
}

export type CellValue = 
  | { stringVal: string; isNull: false }
  | { intVal: number; isNull: false }
  | { doubleVal: number; isNull: false }
  | { boolVal: boolean; isNull: false }
  | { bytesVal: Uint8Array; isNull: false }
  | { isNull: true };

export interface Row {
  cells: Record<string, CellValue>;
}

export interface QueryRowsRequest {
  keyspace: string;
  table: string;
  pageSize: number;
}

export interface QueryRowsResponse {
  rows: Row[];
  cursorId: string;
  hasMore: boolean;
  totalFetched: number;
}

export interface GetNextPageRequest {
  cursorId: string;
}

export interface GetNextPageResponse {
  rows: Row[];
  cursorId: string;
  hasMore: boolean;
}

export interface FilterRowsRequest {
  keyspace: string;
  table: string;
  whereClause: string;
  pageSize: number;
}

export interface FilterRowsResponse {
  rows: Row[];
  cursorId: string;
  hasMore: boolean;
}

export interface ApiError {
  code: string;
  message: string;
  details: Record<string, string>;
}

export interface ViewState {
  keyspace: string;
  table: string;
  filter: string;
  page: number;
}
