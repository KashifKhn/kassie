export interface ProfileInfo {
  name: string;
  hosts: string[];
  port: number;
  keyspace?: string;
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

export type LogoutRequest = Record<string, never>;

export type LogoutResponse = Record<string, never>;

export type GetProfilesRequest = Record<string, never>;

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

export type ListKeyspacesRequest = Record<string, never>;

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
  | { stringVal: string; isNull: false; cqlType?: string }
  | { intVal: number; isNull: false; cqlType?: string }
  | { doubleVal: number; isNull: false; cqlType?: string }
  | { boolVal: boolean; isNull: false; cqlType?: string }
  | { bytesVal: Uint8Array; isNull: false; cqlType?: string }
  | { isNull: true; cqlType?: string };

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

export interface ExecuteQueryRequest {
  cql: string;
  pageSize: number;
  trace?: boolean;
}

export interface ExecuteQueryResponse {
  rows: Row[];
  cursorId: string;
  hasMore: boolean;
  totalFetched: number;
  traceId?: string;
}

export interface QueryHistoryEntry {
  cql: string;
  executedAt: number;
  latencyMs?: number;
}

export interface SlowQuery {
  cql: string;
  lastLatencyMs: number;
  avgLatencyMs: number;
  maxLatencyMs: number;
  execCount: number;
  lastExecutedAt: number;
}

export interface SavedQuery {
  name: string;
  cql: string;
  createdAt: number;
}

export interface TableStats {
  rowCount: number;
  meanPartitionSizeBytes: number;
  maxPartitionSizeBytes: number;
  estimateAvailable: boolean;
}

export interface TraceEvent {
  activity: string;
  source: string;
  elapsedUs: number;
  thread: string;
}

export interface TraceData {
  events: TraceEvent[];
  durationUs: number;
  coordinator: string;
  ready: boolean;
}

export interface ClusterNodeInfo {
  address: string;
  dataCenter: string;
  rack: string;
  releaseVersion: string;
  tokenCount: number;
  local: boolean;
  status: string;
}

export interface AdvisorFinding {
  severity: string;
  rule: string;
  table: string;
  message: string;
  remediation: string;
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
