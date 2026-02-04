package cli

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/KashifKhn/kassie/internal/server"
	"github.com/spf13/cobra"
)

var (
	webPort    int
	noBrowser  bool
	webProfile string
)

func newWebCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Launch web interface",
		Long: `Launch Kassie web interface in your browser.

Starts an embedded server and opens your default browser to access the web UI.`,
		RunE: runWeb,
	}

	cmd.Flags().IntVar(&webPort, "port", 8080, "HTTP port")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "don't auto-open browser")
	cmd.Flags().StringVar(&webProfile, "profile", "", "default profile to connect")

	return cmd
}

func runWeb(cmd *cobra.Command, args []string) error {
	if webProfile != "" {
		if _, err := appConfig.GetProfile(webProfile); err != nil {
			return fmt.Errorf("profile not found: %s", webProfile)
		}
	}

	jwtSecret := os.Getenv("KASSIE_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "web-mode-secret"
	}

	embeddedCfg := &server.EmbeddedServerConfig{
		JWTSecret:      jwtSecret,
		HTTPPort:       webPort,
		AllowedOrigins: []string{"*"},
	}

	embeddedServer, err := server.NewEmbeddedServer(appConfig, embeddedCfg, appLogger)
	if err != nil {
		return fmt.Errorf("failed to create embedded server: %w", err)
	}

	if err := embeddedServer.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	time.Sleep(200 * time.Millisecond)

	url := fmt.Sprintf("http://%s", embeddedServer.HTTPAddress())
	appLogger.With().Str("url", url).Logger().Info("web interface available")

	if !noBrowser && appConfig.Clients.Web.AutoOpenBrowser {
		if err := openBrowser(url); err != nil {
			appLogger.With().Err(err).Logger().Warn("failed to open browser")
		}
	}

	fmt.Printf("\n🚀 Kassie Web UI is running at: %s\n\n", url)
	if webProfile != "" {
		fmt.Printf("Default profile: %s\n", webProfile)
	}
	fmt.Println("Press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n\nShutting down...")

	if err := embeddedServer.Stop(); err != nil {
		appLogger.With().Err(err).Logger().Warn("server shutdown error")
	}

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
