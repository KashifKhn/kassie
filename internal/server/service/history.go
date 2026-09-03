package service

import (
	"context"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/state"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HistoryService struct {
	pb.UnimplementedHistoryServiceServer
	store   SessionStore
	queries *state.QueryStore
}

func NewHistoryService(store SessionStore, queries *state.QueryStore) *HistoryService {
	return &HistoryService{
		store:   store,
		queries: queries,
	}
}

func (h *HistoryService) ListQueryHistory(ctx context.Context, req *pb.ListQueryHistoryRequest) (*pb.ListQueryHistoryResponse, error) {
	session, err := GetSessionFromContext(ctx, h.store)
	if err != nil {
		return nil, err
	}

	limit := int(req.Limit)
	if limit <= 0 || limit > state.MaxHistoryEntries {
		limit = 50
	}

	entries := h.queries.History(session.Profile.Name, limit)

	out := make([]*pb.QueryHistoryEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, &pb.QueryHistoryEntry{
			Cql:        e.CQL,
			ExecutedAt: e.ExecutedAt,
		})
	}

	return &pb.ListQueryHistoryResponse{Entries: out}, nil
}

func (h *HistoryService) ClearQueryHistory(ctx context.Context, req *pb.ClearQueryHistoryRequest) (*pb.ClearQueryHistoryResponse, error) {
	session, err := GetSessionFromContext(ctx, h.store)
	if err != nil {
		return nil, err
	}

	cleared := h.queries.ClearHistory(session.Profile.Name)
	if err := h.queries.Persist(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to persist history: %v", err)
	}

	return &pb.ClearQueryHistoryResponse{Cleared: cleared > 0}, nil
}

func (h *HistoryService) SaveQuery(ctx context.Context, req *pb.SaveQueryRequest) (*pb.SaveQueryResponse, error) {
	session, err := GetSessionFromContext(ctx, h.store)
	if err != nil {
		return nil, err
	}

	if req.Cql == "" {
		return nil, status.Error(codes.InvalidArgument, "cql is required")
	}

	if err := h.queries.SaveQuery(session.Profile.Name, req.Name, req.Cql); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := h.queries.Persist(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to persist saved query: %v", err)
	}

	return &pb.SaveQueryResponse{
		Query: &pb.SavedQuery{Name: req.Name, Cql: req.Cql},
	}, nil
}

func (h *HistoryService) ListSavedQueries(ctx context.Context, req *pb.ListSavedQueriesRequest) (*pb.ListSavedQueriesResponse, error) {
	session, err := GetSessionFromContext(ctx, h.store)
	if err != nil {
		return nil, err
	}

	saved := h.queries.SavedQueries(session.Profile.Name)

	out := make([]*pb.SavedQuery, 0, len(saved))
	for _, sq := range saved {
		out = append(out, &pb.SavedQuery{
			Name:      sq.Name,
			Cql:       sq.CQL,
			CreatedAt: sq.CreatedAt,
		})
	}

	return &pb.ListSavedQueriesResponse{Queries: out}, nil
}

func (h *HistoryService) DeleteSavedQuery(ctx context.Context, req *pb.DeleteSavedQueryRequest) (*pb.DeleteSavedQueryResponse, error) {
	session, err := GetSessionFromContext(ctx, h.store)
	if err != nil {
		return nil, err
	}

	deleted := h.queries.DeleteSavedQuery(session.Profile.Name, req.Name)
	if !deleted {
		return nil, status.Errorf(codes.NotFound, "saved query %q not found", req.Name)
	}
	if err := h.queries.Persist(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to persist deletion: %v", err)
	}

	return &pb.DeleteSavedQueryResponse{Deleted: true}, nil
}
