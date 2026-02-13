package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	cancel      context.CancelFunc
}

func NewEmbeddedServer(appCfg *config.Config, cfg *EmbeddedServerConfig, log *logger.Logger) (*EmbeddedServer, error) {
	if cfg.JWTSecret == "" {
		secret, err := generateRandomSecret(32)
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWT secret: %w", err)
		}
		cfg.JWTSecret = secret
		log.Warn("no JWT secret provided, generated random secret for this session")
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

	ctx, cancel := context.WithCancel(context.Background())

	httpGateway, err := gateway.NewGateway(gatewayCfg, log)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create HTTP gateway: %w", err)
	}

	if err := httpGateway.RegisterServices(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to register gateway services: %w", err)
	}

	return &EmbeddedServer{
		grpcServer:  grpcServer,
		httpGateway: httpGateway,
		cfg:         cfg,
		logger:      log,
		cancel:      cancel,
	}, nil
}

func (e *EmbeddedServer) Start() error {
	if err := e.grpcServer.Listen(); err != nil {
		return fmt.Errorf("failed to start gRPC listener: %w", err)
	}

	go func() {
		if err := e.grpcServer.Serve(); err != nil {
			e.logger.With().Err(err).Logger().Error("gRPC server failed")
		}
	}()

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

func generateRandomSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
