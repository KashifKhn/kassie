package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/KashifKhn/kassie/internal/server/gateway"
	"github.com/KashifKhn/kassie/internal/server/grpc"
	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/KashifKhn/kassie/internal/shared/logger"
)

type EmbeddedServerConfig struct {
	JWTSecret      string
	GRPCPort       int
	HTTPPort       int
	AllowedOrigins []string
}

type EmbeddedServer struct {
	grpcServer  *grpc.Server
	httpGateway *gateway.Gateway
	cfg         *EmbeddedServerConfig
	logger      *logger.Logger
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewEmbeddedServer(appCfg *config.Config, cfg *EmbeddedServerConfig, log *logger.Logger) (*EmbeddedServer, error) {
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "embedded-default-secret"
	}

	if cfg.GRPCPort == 0 {
		port, err := getFreePort()
		if err != nil {
			return nil, fmt.Errorf("failed to find free port for gRPC: %w", err)
		}
		cfg.GRPCPort = port
	}

	if cfg.HTTPPort == 0 {
		port, err := getFreePort()
		if err != nil {
			return nil, fmt.Errorf("failed to find free port for HTTP: %w", err)
		}
		cfg.HTTPPort = port
	}

	grpcCfg := &grpc.ServerConfig{
		Host:      "127.0.0.1",
		Port:      cfg.GRPCPort,
		JWTSecret: cfg.JWTSecret,
	}

	grpcServer, err := grpc.NewServer(grpcCfg, appCfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC server: %w", err)
	}

	gatewayCfg := &gateway.GatewayConfig{
		Host:           "127.0.0.1",
		Port:           cfg.HTTPPort,
		GRPCAddress:    fmt.Sprintf("127.0.0.1:%d", cfg.GRPCPort),
		AllowedOrigins: cfg.AllowedOrigins,
	}

	httpGateway, err := gateway.NewGateway(gatewayCfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP gateway: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &EmbeddedServer{
		grpcServer:  grpcServer,
		httpGateway: httpGateway,
		cfg:         cfg,
		logger:      log,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

func (e *EmbeddedServer) Start() error {
	go func() {
		if err := e.grpcServer.Start(); err != nil {
			e.logger.With().Err(err).Logger().Error("gRPC server failed")
		}
	}()

	time.Sleep(100 * time.Millisecond)

	if err := e.httpGateway.RegisterServices(e.ctx); err != nil {
		return fmt.Errorf("failed to register gateway services: %w", err)
	}

	go func() {
		if err := e.httpGateway.Start(); err != nil {
			e.logger.With().Err(err).Logger().Error("HTTP gateway failed")
		}
	}()

	e.logger.Info("embedded server started")
	return nil
}

func (e *EmbeddedServer) Stop() error {
	e.logger.Info("stopping embedded server")
	e.cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.httpGateway.Stop(shutdownCtx); err != nil {
		e.logger.With().Err(err).Logger().Warn("HTTP gateway shutdown error")
	}

	if err := e.grpcServer.Stop(); err != nil {
		e.logger.With().Err(err).Logger().Warn("gRPC server shutdown error")
	}

	e.logger.Info("embedded server stopped")
	return nil
}

func (e *EmbeddedServer) GRPCAddress() string {
	return e.grpcServer.Address()
}

func (e *EmbeddedServer) HTTPAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", e.cfg.HTTPPort)
}

func getFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
