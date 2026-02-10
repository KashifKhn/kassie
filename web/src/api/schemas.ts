import { z } from 'zod';

export const ProfileInfoSchema = z.object({
  name: z.string(),
  hosts: z.array(z.string()),
  port: z.number(),
  keyspace: z.string(),
  sslEnabled: z.boolean(),
});

export const LoginRequestSchema = z.object({
  profile: z.string(),
});

export const LoginResponseSchema = z.object({
  accessToken: z.string(),
  refreshToken: z.string(),
  expiresAt: z.number(),
  profile: ProfileInfoSchema,
});

export const RefreshRequestSchema = z.object({
  refreshToken: z.string(),
});

export const RefreshResponseSchema = z.object({
  accessToken: z.string(),
  expiresAt: z.number(),
});

export const LogoutRequestSchema = z.object({});

export const LogoutResponseSchema = z.object({});

export const GetProfilesRequestSchema = z.object({});

export const GetProfilesResponseSchema = z.object({
  profiles: z.array(ProfileInfoSchema),
});

export const KeyspaceSchema = z.object({
  name: z.string(),
  replicationStrategy: z.string(),
  replication: z.record(z.string(), z.string()),
});

export const TableSchema = z.object({
  name: z.string(),
  keyspace: z.string(),
  estimatedRows: z.number(),
});

export const ColumnSchema = z.object({
  name: z.string(),
  type: z.string(),
  isPartitionKey: z.boolean(),
  isClusteringKey: z.boolean(),
  position: z.number(),
});

export const TableSchemaSchema = z.object({
  keyspace: z.string(),
  table: z.string(),
  columns: z.array(ColumnSchema),
  partitionKeys: z.array(z.string()),
  clusteringKeys: z.array(z.string()),
});

export const ListKeyspacesRequestSchema = z.object({});

export const ListKeyspacesResponseSchema = z.object({
  keyspaces: z.array(KeyspaceSchema),
});

export const ListTablesRequestSchema = z.object({
  keyspace: z.string(),
});

export const ListTablesResponseSchema = z.object({
  tables: z.array(TableSchema),
});

export const GetTableSchemaRequestSchema = z.object({
  keyspace: z.string(),
  table: z.string(),
});

export const GetTableSchemaResponseSchema = z.object({
  schema: TableSchemaSchema,
});

export const CellValueSchema = z.union([
  z.object({ stringVal: z.string(), isNull: z.literal(false) }),
  z.object({ intVal: z.number(), isNull: z.literal(false) }),
  z.object({ doubleVal: z.number(), isNull: z.literal(false) }),
  z.object({ boolVal: z.boolean(), isNull: z.literal(false) }),
  z.object({ bytesVal: z.instanceof(Uint8Array), isNull: z.literal(false) }),
  z.object({ isNull: z.literal(true) }),
]);

export const RowSchema = z.object({
  cells: z.record(z.string(), CellValueSchema),
});

export const QueryRowsRequestSchema = z.object({
  keyspace: z.string(),
  table: z.string(),
  pageSize: z.number(),
});

export const QueryRowsResponseSchema = z.object({
  rows: z.array(RowSchema),
  cursorId: z.string(),
  hasMore: z.boolean(),
  totalFetched: z.number(),
});

export const GetNextPageRequestSchema = z.object({
  cursorId: z.string(),
});

export const GetNextPageResponseSchema = z.object({
  rows: z.array(RowSchema),
  cursorId: z.string(),
  hasMore: z.boolean(),
});

export const FilterRowsRequestSchema = z.object({
  keyspace: z.string(),
  table: z.string(),
  whereClause: z.string(),
  pageSize: z.number(),
});

export const FilterRowsResponseSchema = z.object({
  rows: z.array(RowSchema),
  cursorId: z.string(),
  hasMore: z.boolean(),
});

export const ApiErrorSchema = z.object({
  code: z.string(),
  message: z.string(),
  details: z.record(z.string(), z.string()),
});

export const ViewStateSchema = z.object({
  keyspace: z.string(),
  table: z.string(),
  filter: z.string(),
  page: z.number(),
});
