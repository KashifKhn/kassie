# API Reference

Kassie exposes both gRPC and REST APIs for programmatic access.

## gRPC Services

### SessionService

```protobuf
service SessionService {
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc Refresh(RefreshRequest) returns (RefreshResponse);
  rpc Logout(LogoutRequest) returns (LogoutResponse);
  rpc GetProfiles(GetProfilesRequest) returns (GetProfilesResponse);
}
```

### SchemaService

```protobuf
service SchemaService {
  rpc ListKeyspaces(ListKeyspacesRequest) returns (ListKeyspacesResponse);
  rpc ListTables(ListTablesRequest) returns (ListTablesResponse);
  rpc GetTableSchema(GetTableSchemaRequest) returns (GetTableSchemaResponse);
}
```

### DataService

```protobuf
service DataService {
  rpc QueryRows(QueryRowsRequest) returns (QueryRowsResponse);
  rpc GetNextPage(GetNextPageRequest) returns (GetNextPageResponse);
  rpc FilterRows(FilterRowsRequest) returns (FilterRowsResponse);
}
```

## REST Endpoints

All gRPC services are exposed via HTTP using grpc-gateway.

See the [Architecture documentation](/architecture/protocol) for detailed API information.
