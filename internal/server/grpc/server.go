package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/db"
	"github.com/KashifKhn/kassie/internal/server/service"
	"github.com/KashifKhn/kassie/internal/server/state"
	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/KashifKhn/kassie/internal/shared/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type ServerConfig struct {
	Host      string
	Port      int
	JWTSecret string
}

type Server struct {
	cfg            *ServerConfig
	grpcServer     *grpc.Server
	sessionService *service.SessionService
	schemaService  *service.SchemaService
	dataService    *service.DataService
	listener       net.Listener
	logger         *logger.Logger
}

func NewServer(cfg *ServerConfig, appCfg *config.Config, log *logger.Logger) (*Server, error) {
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "default-secret-change-in-production"
	}

	pool := db.NewPool()
	store := state.NewStore(7 * 24 * time.Hour)
	auth := service.NewAuthService(cfg.JWTSecret)

	sessionSvc := service.NewSessionService(appCfg, pool, store, auth)
	schemaSvc := service.NewSchemaService(store)
	dataSvc := service.NewDataService(store)

	unaryInterceptor := NewAuthInterceptor(auth, store, log)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
	)

	pb.RegisterSessionServiceServer(grpcServer, sessionSvc)
	pb.RegisterSchemaServiceServer(grpcServer, schemaSvc)
	pb.RegisterDataServiceServer(grpcServer, dataSvc)

	reflection.Register(grpcServer)

	s := &Server{
		cfg:            cfg,
		grpcServer:     grpcServer,
		sessionService: sessionSvc,
		schemaService:  schemaSvc,
		dataService:    dataSvc,
		logger:         log,
	}

	return s, nil
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	s.logger.With().Str("address", addr).Logger().Info("gRPC server listening")

	if err := s.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) Stop() error {
	s.logger.Info("stopping gRPC server")

	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		s.logger.Info("gRPC server stopped gracefully")
	case <-time.After(10 * time.Second):
		s.grpcServer.Stop()
		s.logger.Warn("gRPC server force stopped after timeout")
	}

	if s.listener != nil {
		s.listener.Close()
	}

	return nil
}

func (s *Server) Address() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
}

func (s *Server) ServeAsync(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if err := s.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(100 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return s.Stop()
	}
}
