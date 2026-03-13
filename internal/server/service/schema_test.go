package service

import (
	"context"
	"testing"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/db"
	"github.com/KashifKhn/kassie/internal/server/state"
	"github.com/KashifKhn/kassie/internal/shared/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockSchemaStore struct {
	session *state.Session
	err     error
}

func (m *mockSchemaStore) Create(id string, profile *config.Profile, conn *db.Session) *state.Session {
	return nil
}

func (m *mockSchemaStore) Get(id string) (*state.Session, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.session, nil
}

func (m *mockSchemaStore) Delete(id string) {
}

func (m *mockSchemaStore) CloseAll() {
}

func (m *mockSchemaStore) Close() {
}

func TestSchemaService_ListTables_MissingKeyspace(t *testing.T) {
	store := &mockSchemaStore{}
	service := NewSchemaService(store)

	_, err := service.ListTables(context.Background(), &pb.ListTablesRequest{Keyspace: ""})

	if err == nil {
		t.Fatal("expected error for missing keyspace")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestSchemaService_GetTableSchema_MissingKeyspace(t *testing.T) {
	store := &mockSchemaStore{}
	service := NewSchemaService(store)

	_, err := service.GetTableSchema(context.Background(), &pb.GetTableSchemaRequest{
		Keyspace: "",
		Table:    "users",
	})

	if err == nil {
		t.Fatal("expected error for missing keyspace")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}

func TestSchemaService_GetTableSchema_MissingTable(t *testing.T) {
	store := &mockSchemaStore{}
	service := NewSchemaService(store)

	_, err := service.GetTableSchema(context.Background(), &pb.GetTableSchemaRequest{
		Keyspace: "users_ks",
		Table:    "",
	})

	if err == nil {
		t.Fatal("expected error for missing table")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", st.Code())
	}
}
