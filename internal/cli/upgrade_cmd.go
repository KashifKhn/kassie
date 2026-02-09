package cli

import (
	"github.com/KashifKhn/kassie/internal/cli/upgrade"
	"github.com/spf13/cobra"
)

func newUpgradeCmd() *cobra.Command {
	return upgrade.NewCommand()
}
