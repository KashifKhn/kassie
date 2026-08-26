package grpc_test

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/client"
	servergrpc "github.com/KashifKhn/kassie/internal/server/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/test/bufconn"
)

const payloadKB = 6 * 1024

type stubSessionService struct {
	pb.UnimplementedSessionServiceServer
}

func (s *stubSessionService) GetProfiles(ctx context.Context, req *pb.GetProfilesRequest) (*pb.GetProfilesResponse, error) {
	resp := &pb.GetProfilesResponse{}
	for range 1024 {
		resp.Profiles = append(resp.Profiles, &pb.ProfileInfo{
			Name:     "profile-" + strings.Repeat("x", payloadKB),
			Hosts:    []string{"127.0.0.1"},
			Port:     9042,
			Keyspace: "ks",
		})
	}
	return resp, nil
}

type recordingStats struct {
	mu         sync.Mutex
	inPayloads []*stats.InPayload
}

func (r *recordingStats) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

func (r *recordingStats) HandleRPC(_ context.Context, s stats.RPCStats) {
	if p, ok := s.(*stats.InPayload); ok {
		r.mu.Lock()
		r.inPayloads = append(r.inPayloads, p)
		r.mu.Unlock()
	}
}

func (r *recordingStats) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

func (r *recordingStats) HandleConn(context.Context, stats.ConnStats) {}

func startStubServer(t *testing.T) (*recordingStats, pb.SessionServiceClient) {
	t.Helper()

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer(servergrpc.ServerOptions(nil, nil)...)
	pb.RegisterSessionServiceServer(srv, &stubSessionService{})

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	statsHandler := &recordingStats{}
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		[]grpc.DialOption{
			grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
				return lis.DialContext(ctx)
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithStatsHandler(statsHandler),
			grpc.WithDefaultCallOptions(client.DefaultCallOptions()...),
		}...,
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return statsHandler, pb.NewSessionServiceClient(conn)
}

func TestGzipCompressionRoundTrip(t *testing.T) {
	statsHandler, client := startStubServer(t)

	resp, err := client.GetProfiles(context.Background(), &pb.GetProfilesRequest{})
	if err != nil {
		t.Fatalf("GetProfiles with gzip compression failed: %v", err)
	}
	if len(resp.Profiles) == 0 {
		t.Fatal("expected profiles in response")
	}

	statsHandler.mu.Lock()
	defer statsHandler.mu.Unlock()

	var wireBytes, rawBytes int64
	for _, p := range statsHandler.inPayloads {
		wireBytes += int64(p.CompressedLength)
		rawBytes += int64(p.Length)
	}

	if rawBytes < 6*1024*1024 {
		t.Fatalf("payload too small to exercise limits: %d bytes", rawBytes)
	}
	if wireBytes >= rawBytes/2 {
		t.Fatalf("response not compressed: %d bytes on wire for %d bytes of data", wireBytes, rawBytes)
	}
}
