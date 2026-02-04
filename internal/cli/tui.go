package cli

import (
	"fmt"

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
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║          Kassie TUI - Coming Soon! 🚧               ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("The Terminal UI is currently under development.")
	fmt.Println("In the meantime, try the web interface:")
	fmt.Println()
	fmt.Println("  kassie web")
	fmt.Println()
	fmt.Println("Or run as a standalone server:")
	fmt.Println()
	fmt.Println("  kassie server")
	fmt.Println()

	return nil
}
