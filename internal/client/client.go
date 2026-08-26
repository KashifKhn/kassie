package client

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	session pb.SessionServiceClient
	schema  pb.SchemaServiceClient
	data    pb.DataServiceClient
	history pb.HistoryServiceClient

	mu           sync.RWMutex
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	profile      string
}

func New(addr string) (*Client, error) {
	c := &Client{}

	conn, err := grpc.NewClient(addr, DialOptions(c.authInterceptor(), c.streamAuthInterceptor())...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.session = pb.NewSessionServiceClient(conn)
	c.schema = pb.NewSchemaServiceClient(conn)
	c.data = pb.NewDataServiceClient(conn)
	c.history = pb.NewHistoryServiceClient(conn)

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

func (c *Client) GetTableStats(ctx context.Context, keyspace, table string) (*pb.TableStats, error) {
	resp, err := c.schema.GetTableStats(ctx, &pb.GetTableStatsRequest{
		Keyspace: keyspace,
		Table:    table,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get table stats: %w", err)
	}
	return resp.Stats, nil
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

func (c *Client) ExportRows(ctx context.Context, keyspace, table, where string, format pb.ExportFormat, fetchSize int32, onChunk func(*pb.ExportChunk) error) error {
	stream, err := c.data.ExportRows(ctx, &pb.ExportRowsRequest{
		Keyspace:    keyspace,
		Table:       table,
		WhereClause: where,
		Format:      format,
		FetchSize:   fetchSize,
	})
	if err != nil {
		return fmt.Errorf("failed to start export: %w", err)
	}

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
		if err := onChunk(chunk); err != nil {
			return err
		}
		if chunk.Done {
			return nil
		}
	}
}

func (c *Client) ExecuteQuery(ctx context.Context, cql string, pageSize int32) (*pb.ExecuteQueryResponse, error) {
	resp, err := c.data.ExecuteQuery(ctx, &pb.ExecuteQueryRequest{
		Cql:      cql,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return resp, nil
}

func (c *Client) ListQueryHistory(ctx context.Context, limit int32) ([]*pb.QueryHistoryEntry, error) {
	resp, err := c.history.ListQueryHistory(ctx, &pb.ListQueryHistoryRequest{Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("failed to list query history: %w", err)
	}
	return resp.Entries, nil
}

func (c *Client) ClearQueryHistory(ctx context.Context) error {
	if _, err := c.history.ClearQueryHistory(ctx, &pb.ClearQueryHistoryRequest{}); err != nil {
		return fmt.Errorf("failed to clear query history: %w", err)
	}
	return nil
}

func (c *Client) SaveQuery(ctx context.Context, name, cql string) (*pb.SavedQuery, error) {
	resp, err := c.history.SaveQuery(ctx, &pb.SaveQueryRequest{Name: name, Cql: cql})
	if err != nil {
		return nil, fmt.Errorf("failed to save query: %w", err)
	}
	return resp.Query, nil
}

func (c *Client) ListSavedQueries(ctx context.Context) ([]*pb.SavedQuery, error) {
	resp, err := c.history.ListSavedQueries(ctx, &pb.ListSavedQueriesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list saved queries: %w", err)
	}
	return resp.Queries, nil
}

func (c *Client) DeleteSavedQuery(ctx context.Context, name string) error {
	if _, err := c.history.DeleteSavedQuery(ctx, &pb.DeleteSavedQueryRequest{Name: name}); err != nil {
		return fmt.Errorf("failed to delete saved query: %w", err)
	}
	return nil
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
