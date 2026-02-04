package service

import (
	"context"
	"fmt"
	"sort"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/state"
)

type SchemaService struct {
	pb.UnimplementedSchemaServiceServer
	store *state.Store
}

func NewSchemaService(store *state.Store) *SchemaService {
	return &SchemaService{
		store: store,
	}
}

func (s *SchemaService) ListKeyspaces(ctx context.Context, req *pb.ListKeyspacesRequest) (*pb.ListKeyspacesResponse, error) {
	session, err := GetSessionFromContext(ctx, s.store)
	if err != nil {
		return nil, err
	}

	query := `SELECT keyspace_name, replication FROM system_schema.keyspaces`
	rows, err := session.Connection.FetchAll(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keyspaces: %w", err)
	}

	keyspaces := make([]*pb.Keyspace, 0, len(rows))
	for _, row := range rows {
		name, _ := row["keyspace_name"].(string)
		replication, _ := row["replication"].(map[string]string)

		if name == "" {
			continue
		}

		ks := &pb.Keyspace{
			Name:        name,
			Replication: replication,
		}

		if replication != nil {
			if strategy, ok := replication["class"]; ok {
				ks.ReplicationStrategy = strategy
			}
		}

		keyspaces = append(keyspaces, ks)
	}

	sort.Slice(keyspaces, func(i, j int) bool {
		return keyspaces[i].Name < keyspaces[j].Name
	})

	return &pb.ListKeyspacesResponse{
		Keyspaces: keyspaces,
	}, nil
}

func (s *SchemaService) ListTables(ctx context.Context, req *pb.ListTablesRequest) (*pb.ListTablesResponse, error) {
	if req.Keyspace == "" {
		return nil, fmt.Errorf("keyspace is required")
	}

	session, err := GetSessionFromContext(ctx, s.store)
	if err != nil {
		return nil, err
	}

	query := `SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?`
	rows, err := session.Connection.FetchAll(ctx, query, req.Keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tables: %w", err)
	}

	tables := make([]*pb.Table, 0, len(rows))
	for _, row := range rows {
		name, ok := row["table_name"].(string)
		if !ok || name == "" {
			continue
		}

		tables = append(tables, &pb.Table{
			Name:     name,
			Keyspace: req.Keyspace,
		})
	}

	sort.Slice(tables, func(i, j int) bool {
		return tables[i].Name < tables[j].Name
	})

	return &pb.ListTablesResponse{
		Tables: tables,
	}, nil
}

func (s *SchemaService) GetTableSchema(ctx context.Context, req *pb.GetTableSchemaRequest) (*pb.GetTableSchemaResponse, error) {
	if req.Keyspace == "" || req.Table == "" {
		return nil, fmt.Errorf("keyspace and table are required")
	}

	session, err := GetSessionFromContext(ctx, s.store)
	if err != nil {
		return nil, err
	}

	query := `SELECT column_name, type, kind, position FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ?`
	rows, err := session.Connection.FetchAll(ctx, query, req.Keyspace, req.Table)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch table schema: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("table not found: %s.%s", req.Keyspace, req.Table)
	}

	columns := make([]*pb.Column, 0, len(rows))
	partitionKeys := make([]string, 0)
	clusteringKeys := make([]string, 0)

	for _, row := range rows {
		colName, _ := row["column_name"].(string)
		colType, _ := row["type"].(string)
		kind, _ := row["kind"].(string)
		position, _ := row["position"].(int)

		if colName == "" {
			continue
		}

		isPartition := kind == "partition_key"
		isClustering := kind == "clustering"

		col := &pb.Column{
			Name:            colName,
			Type:            colType,
			IsPartitionKey:  isPartition,
			IsClusteringKey: isClustering,
			Position:        int32(position),
		}

		columns = append(columns, col)

		if isPartition {
			partitionKeys = append(partitionKeys, colName)
		}
		if isClustering {
			clusteringKeys = append(clusteringKeys, colName)
		}
	}

	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Position < columns[j].Position
	})

	return &pb.GetTableSchemaResponse{
		Schema: &pb.TableSchema{
			Keyspace:       req.Keyspace,
			Table:          req.Table,
			Columns:        columns,
			PartitionKeys:  partitionKeys,
			ClusteringKeys: clusteringKeys,
		},
	}, nil
}
