package service

import (
	"context"
	"fmt"
	"net"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SchemaService) GetClusterInfo(ctx context.Context, req *pb.GetClusterInfoRequest) (*pb.GetClusterInfoResponse, error) {
	session, err := GetSessionFromContext(ctx, s.store)
	if err != nil {
		return nil, err
	}

	nodes := []*pb.ClusterNode{}

	local, err := session.Connection.FetchAll(ctx,
		`SELECT rpc_address, data_center, rack, release_version, tokens FROM system.local`)
	if err == nil {
		for _, row := range local {
			nodes = append(nodes, clusterNodeFromRow(row, true))
		}
	}

	peers, err := session.Connection.FetchAll(ctx,
		`SELECT rpc_address, data_center, rack, release_version, tokens FROM system.peers`)
	if err == nil {
		for _, row := range peers {
			nodes = append(nodes, clusterNodeFromRow(row, false))
		}
	}

	if len(nodes) == 0 {
		return nil, status.Errorf(codes.Internal, "failed to read cluster topology")
	}

	return &pb.GetClusterInfoResponse{Nodes: nodes}, nil
}

func clusterNodeFromRow(row map[string]interface{}, local bool) *pb.ClusterNode {
	node := &pb.ClusterNode{
		Address:        stringValue(row["rpc_address"]),
		DataCenter:     stringValue(row["data_center"]),
		Rack:           stringValue(row["rack"]),
		ReleaseVersion: stringValue(row["release_version"]),
		Local:          local,
		Status:         "up",
		TokenCount:     int32(tokenCount(row["tokens"])),
	}

	if local {
		node.Status = "local"
	}

	return node
}

func stringValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case net.IP:
		return val.String()
	default:
		if v == nil {
			return ""
		}
		return fmt.Sprintf("%v", v)
	}
}

func tokenCount(v interface{}) int {
	switch tokens := v.(type) {
	case []string:
		return len(tokens)
	case []interface{}:
		return len(tokens)
	case map[string]interface{}:
		return len(tokens)
	default:
		return 0
	}
}
