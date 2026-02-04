package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KashifKhn/kassie/internal/server/gateway"
	"github.com/KashifKhn/kassie/internal/server/grpc"
	"github.com/spf13/cobra"
)

var (
	grpcPort int
	httpPort int
	bindHost string
)

func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run standalone server",
		Long: `Run Kassie as a standalone server exposing both gRPC and HTTP/REST APIs.

This mode is suitable for team environments where multiple clients need to connect
to a shared Kassie instance.`,
		RunE: runServer,
	}

	cmd.Flags().IntVar(&grpcPort, "grpc-port", 50051, "gRPC server port")
	cmd.Flags().IntVar(&httpPort, "http-port", 8080, "HTTP gateway port")
	cmd.Flags().StringVar(&bindHost, "host", "0.0.0.0", "bind address")

	return cmd
}

func runServer(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jwtSecret := os.Getenv("KASSIE_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-this-secret-in-production"
		appLogger.Warn("using default JWT secret, set KASSIE_JWT_SECRET env variable")
	}

	grpcCfg := &grpc.ServerConfig{
		Host:      bindHost,
		Port:      grpcPort,
		JWTSecret: jwtSecret,
	}

	grpcServer, err := grpc.NewServer(grpcCfg, appConfig, appLogger)
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}

	go func() {
		if err := grpcServer.Start(); err != nil {
			appLogger.With().Err(err).Logger().Error("gRPC server failed")
			cancel()
		}
	}()

	gatewayCfg := &gateway.GatewayConfig{
		Host:           bindHost,
		Port:           httpPort,
		GRPCAddress:    fmt.Sprintf("%s:%d", "127.0.0.1", grpcPort),
		AllowedOrigins: []string{"*"},
	}

	httpGateway, err := gateway.NewGateway(gatewayCfg, appLogger)
	if err != nil {
		return fmt.Errorf("failed to create HTTP gateway: %w", err)
	}

	if err := httpGateway.RegisterServices(ctx); err != nil {
		return fmt.Errorf("failed to register services: %w", err)
	}

	go func() {
		if err := httpGateway.Start(); err != nil {
			appLogger.With().Err(err).Logger().Error("HTTP gateway failed")
			cancel()
		}
	}()

	appLogger.With().
		Str("grpc_address", grpcServer.Address()).
		Str("http_address", fmt.Sprintf("%s:%d", bindHost, httpPort)).
		Logger().Info("server started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		appLogger.Info("shutting down server")
	case <-ctx.Done():
		appLogger.Info("server context cancelled")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpGateway.Stop(shutdownCtx); err != nil {
		appLogger.With().Err(err).Logger().Warn("HTTP gateway shutdown error")
	}

	if err := grpcServer.Stop(); err != nil {
		appLogger.With().Err(err).Logger().Warn("gRPC server shutdown error")
	}

	appLogger.Info("server stopped")
	return nil
}
