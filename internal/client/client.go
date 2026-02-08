package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	session pb.SessionServiceClient
	schema  pb.SchemaServiceClient
	data    pb.DataServiceClient

	mu           sync.RWMutex
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	profile      string
}

func New(addr string) (*Client, error) {
	c := &Client{}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(c.authInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.session = pb.NewSessionServiceClient(conn)
	c.schema = pb.NewSchemaServiceClient(conn)
	c.data = pb.NewDataServiceClient(conn)

	return c, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) GetProfiles(ctx context.Context) ([]*pb.ProfileInfo, error) {
	resp, err := c.session.GetProfiles(ctx, &pb.GetProfilesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles: %w", err)
	}
	return resp.Profiles, nil
}

func (c *Client) Login(ctx context.Context, profile string) (*pb.ProfileInfo, error) {
	resp, err := c.session.Login(ctx, &pb.LoginRequest{Profile: profile})
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	c.mu.Lock()
	c.accessToken = resp.AccessToken
	c.refreshToken = resp.RefreshToken
	c.expiresAt = time.Unix(resp.ExpiresAt, 0)
	c.profile = profile
	c.mu.Unlock()

	return resp.Profile, nil
}

func (c *Client) Logout(ctx context.Context) error {
	_, err := c.session.Logout(ctx, &pb.LogoutRequest{})

	c.mu.Lock()
	c.accessToken = ""
	c.refreshToken = ""
	c.expiresAt = time.Time{}
	c.profile = ""
	c.mu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}
	return nil
}

func (c *Client) Refresh(ctx context.Context) error {
	c.mu.RLock()
	rt := c.refreshToken
	c.mu.RUnlock()

	if rt == "" {
		return fmt.Errorf("no refresh token available")
	}

	resp, err := c.session.Refresh(ctx, &pb.RefreshRequest{RefreshToken: rt})
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	c.mu.Lock()
	c.accessToken = resp.AccessToken
	c.expiresAt = time.Unix(resp.ExpiresAt, 0)
	c.mu.Unlock()

	return nil
}

func (c *Client) ListKeyspaces(ctx context.Context) ([]*pb.Keyspace, error) {
	resp, err := c.schema.ListKeyspaces(ctx, &pb.ListKeyspacesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list keyspaces: %w", err)
	}
	return resp.Keyspaces, nil
}

func (c *Client) ListTables(ctx context.Context, keyspace string) ([]*pb.Table, error) {
	resp, err := c.schema.ListTables(ctx, &pb.ListTablesRequest{Keyspace: keyspace})
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	return resp.Tables, nil
}

func (c *Client) GetTableSchema(ctx context.Context, keyspace, table string) (*pb.TableSchema, error) {
	resp, err := c.schema.GetTableSchema(ctx, &pb.GetTableSchemaRequest{
		Keyspace: keyspace,
		Table:    table,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get table schema: %w", err)
	}
	return resp.Schema, nil
}

func (c *Client) QueryRows(ctx context.Context, keyspace, table string, pageSize int32) (*pb.QueryRowsResponse, error) {
	resp, err := c.data.QueryRows(ctx, &pb.QueryRowsRequest{
		Keyspace: keyspace,
		Table:    table,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}
	return resp, nil
}

func (c *Client) GetNextPage(ctx context.Context, cursorID string) (*pb.GetNextPageResponse, error) {
	resp, err := c.data.GetNextPage(ctx, &pb.GetNextPageRequest{CursorId: cursorID})
	if err != nil {
		return nil, fmt.Errorf("failed to get next page: %w", err)
	}
	return resp, nil
}

func (c *Client) FilterRows(ctx context.Context, keyspace, table, where string, pageSize int32) (*pb.FilterRowsResponse, error) {
	resp, err := c.data.FilterRows(ctx, &pb.FilterRowsRequest{
		Keyspace:    keyspace,
		Table:       table,
		WhereClause: where,
		PageSize:    pageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to filter rows: %w", err)
	}
	return resp, nil
}

func (c *Client) Profile() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.profile
}

func (c *Client) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken != ""
}
