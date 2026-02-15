package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/KashifKhn/kassie/internal/server/db"
	"github.com/KashifKhn/kassie/internal/server/gateway"
	"github.com/KashifKhn/kassie/internal/server/grpc"
	"github.com/KashifKhn/kassie/internal/server/state"
	"github.com/KashifKhn/kassie/internal/server/web"
	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/spf13/cobra"
)

var (
	webPort    int
	apiPort    int
	noBrowser  bool
	webProfile string
)

func newWebCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Launch web interface",
		Long: `Launch Kassie web interface in your browser.

Starts embedded gRPC server, HTTP API gateway, and web UI server.`,
		RunE: runWeb,
	}

	cmd.Flags().IntVar(&webPort, "web-port", config.DefaultWebPort, "web UI port")
	cmd.Flags().IntVar(&apiPort, "api-port", config.DefaultHTTPPort, "API gateway port")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "don't auto-open browser")
	cmd.Flags().StringVar(&webProfile, "profile", "", "default profile to connect")

	return cmd
}

func runWeb(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if webProfile != "" {
		if _, err := appConfig.GetProfile(webProfile); err != nil {
			return fmt.Errorf("profile not found: %s", webProfile)
		}
	}

	jwtSecret := os.Getenv("KASSIE_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = generateSecret()
		appLogger.Warn("no KASSIE_JWT_SECRET set, generated random secret for this session")
	}

	grpcPort := apiPort - 1

	grpcCfg := &grpc.ServerConfig{
		Host:      config.DefaultHost,
		Port:      grpcPort,
		JWTSecret: jwtSecret,
	}

	pool := db.NewPool()
	store := state.NewStore(config.DefaultSessionTTL)

	grpcDeps := &grpc.ServerDeps{
		Config: appConfig,
		Pool:   pool,
		Store:  store,
	}

	grpcServer, err := grpc.NewServer(grpcCfg, grpcDeps, appLogger)
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}

	if err := grpcServer.Listen(); err != nil {
		return fmt.Errorf("failed to start gRPC listener: %w", err)
	}

	go func() {
		if err := grpcServer.Serve(); err != nil {
			appLogger.With().Err(err).Logger().Error("gRPC server failed")
			cancel()
		}
	}()

	gatewayCfg := &gateway.GatewayConfig{
		Host:           config.DefaultHost,
		Port:           apiPort,
		GRPCAddress:    fmt.Sprintf("%s:%d", config.DefaultHost, grpcPort),
		AllowedOrigins: []string{"*"},
	}

	httpGateway, err := gateway.NewGateway(gatewayCfg, appLogger)
	if err != nil {
		return fmt.Errorf("failed to create HTTP gateway: %w", err)
	}

	if err := httpGateway.RegisterServices(ctx); err != nil {
		return fmt.Errorf("failed to register gateway services: %w", err)
	}

	go func() {
		if err := httpGateway.Start(); err != nil {
			appLogger.With().Err(err).Logger().Error("HTTP gateway failed")
			cancel()
		}
	}()

	webCfg := &web.ServerConfig{
		Host:           config.DefaultHost,
		Port:           webPort,
		AllowedOrigins: []string{"*"},
	}

	webServer, err := web.NewServer(webCfg, appLogger)
	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}

	go func() {
		if err := webServer.Start(); err != nil {
			appLogger.With().Err(err).Logger().Error("web server failed")
			cancel()
		}
	}()

	webURL := fmt.Sprintf("http://%s:%d", config.DefaultHost, webPort)
	apiURL := fmt.Sprintf("http://%s:%d", config.DefaultHost, apiPort)

	appLogger.With().
		Str("web_url", webURL).
		Str("api_url", apiURL).
		Logger().Info("kassie web started")

	if !noBrowser && appConfig.Clients.Web.AutoOpenBrowser {
		if err := openBrowser(webURL); err != nil {
			appLogger.With().Err(err).Logger().Warn("failed to open browser")
		}
	}

	fmt.Printf("\n🚀 Kassie Web UI is running!\n\n")
	fmt.Printf("  Web UI:  %s\n", webURL)
	fmt.Printf("  API:     %s\n\n", apiURL)
	if webProfile != "" {
		fmt.Printf("Default profile: %s\n\n", webProfile)
	}
	fmt.Println("Press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		fmt.Println("\n\nShutting down...")
	case <-ctx.Done():
		fmt.Println("\n\nServer error, shutting down...")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.DefaultShutdownTime)
	defer shutdownCancel()

	if err := webServer.Stop(shutdownCtx); err != nil {
		appLogger.With().Err(err).Logger().Warn("web server shutdown error")
	}

	if err := httpGateway.Stop(shutdownCtx); err != nil {
		appLogger.With().Err(err).Logger().Warn("HTTP gateway shutdown error")
	}

	if err := grpcServer.Stop(); err != nil {
		appLogger.With().Err(err).Logger().Warn("gRPC server shutdown error")
	}

	appLogger.Info("kassie web stopped")
	return nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
