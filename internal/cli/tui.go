package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/server"
	"github.com/KashifKhn/kassie/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	tuiProfile string
	tuiServer  string
)

func newTUICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch terminal user interface",
		Long: `Launch Kassie terminal user interface (TUI).

Provides a keyboard-driven terminal interface for exploring your database.
Coming soon in Phase 4!`,
		RunE: runTUI,
	}

	cmd.Flags().StringVar(&tuiProfile, "profile", "", "profile to connect to")
	cmd.Flags().StringVar(&tuiServer, "server", "", "remote server address (bypasses embedded server)")

	return cmd
}

func runTUI(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var grpcAddr string
	var embedded *server.EmbeddedServer

	if tuiServer != "" {
		grpcAddr = tuiServer
		if tuiProfile != "" {
			appLogger.With().Str("profile", tuiProfile).Logger().Warn("profile flag ignored when using --server")
		}
	} else {
		jwtSecret := os.Getenv("KASSIE_JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "tui-mode-secret"
		}

		embeddedCfg := &server.EmbeddedServerConfig{
			JWTSecret: jwtSecret,
			GRPCPort:  0,
			HTTPPort:  0,
		}

		var err error
		embedded, err = server.NewEmbeddedServer(appConfig, embeddedCfg, appLogger)
		if err != nil {
			return fmt.Errorf("failed to create embedded server: %w", err)
		}

		if err := embedded.Start(); err != nil {
			return fmt.Errorf("failed to start embedded server: %w", err)
		}

		time.Sleep(150 * time.Millisecond)
		grpcAddr = embedded.GRPCAddress()
	}

	clientConn, err := client.New(grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer clientConn.Close()

	if tuiProfile != "" {
		ctxLogin, cancelLogin := context.WithTimeout(context.Background(), 10*time.Second)
		if _, err := clientConn.Login(ctxLogin, tuiProfile); err != nil {
			cancelLogin()
			return fmt.Errorf("failed to login with profile %s: %w", tuiProfile, err)
		}
		cancelLogin()
	}

	program := tea.NewProgram(tui.NewApp(clientConn), tea.WithAltScreen())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			cancel()
			program.Quit()
		case <-ctx.Done():
			program.Quit()
		}
	}()

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	if embedded != nil {
		if err := embedded.Stop(); err != nil {
			appLogger.With().Err(err).Logger().Warn("embedded server shutdown error")
		}
	}

	return nil
}
