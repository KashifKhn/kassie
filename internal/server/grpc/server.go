package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/service"
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

type ServerDeps struct {
	Config service.ProfileProvider
	Pool   service.ConnectionPool
	Store  service.SessionStore
}

func NewServer(cfg *ServerConfig, deps *ServerDeps, log *logger.Logger) (*Server, error) {
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT secret is required")
	}

	auth := service.NewAuthService(cfg.JWTSecret)

	sessionSvc := service.NewSessionService(deps.Config, deps.Pool, deps.Store, auth)
	schemaSvc := service.NewSchemaService(deps.Store)
	dataSvc := service.NewDataService(deps.Store)

	unaryInterceptor := NewAuthInterceptor(auth, deps.Store, log)

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
	if err := s.Listen(); err != nil {
		return err
	}
	return s.Serve()
}

func (s *Server) Listen() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	s.logger.With().Str("address", addr).Logger().Info("gRPC server listening")
	return nil
}

func (s *Server) Serve() error {
	if err := s.grpcServer.Serve(s.listener); err != nil {
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
		_ = s.listener.Close()
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
	if err := s.Listen(); err != nil {
		return err
	}

	errChan := make(chan error, 1)
	go func() {
		if err := s.Serve(); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return s.Stop()
	default:
		return nil
	}
}
