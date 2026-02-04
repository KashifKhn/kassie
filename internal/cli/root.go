package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/KashifKhn/kassie/internal/shared/logger"
	"github.com/KashifKhn/kassie/internal/shared/version"
	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	profile   string
	logLevel  string
	appConfig *config.Config
	appLogger *logger.Logger
)

func NewRootCmd() *cobra.Command {
	var showVersion bool

	cmd := &cobra.Command{
		Use:   "kassie",
		Short: "Database Explorer for Cassandra & ScyllaDB",
		Long: `Kassie - A modern terminal and web explorer for Apache Cassandra and ScyllaDB.

Provides both TUI (Terminal UI) and Web interfaces for exploring and querying
your Cassandra/ScyllaDB clusters with ease.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Printf("Kassie v%s\n", version.Version)
				fmt.Printf("Commit: %s\n", version.Commit)
				fmt.Printf("Built: %s\n", version.BuildDate)
				return nil
			}
			return cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				return nil
			}
			return initConfig()
		},
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/kassie/config.json)")
	cmd.PersistentFlags().StringVar(&profile, "profile", "", "database profile to use")
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	cmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "print version information")

	cmd.AddCommand(newServerCmd())
	cmd.AddCommand(newWebCmd())
	cmd.AddCommand(newTUICmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Kassie v%s\n", version.Version)
			fmt.Printf("Commit: %s\n", version.Commit)
			fmt.Printf("Built: %s\n", version.BuildDate)
		},
	}
}

func initConfig() error {
	level, err := logger.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	appLogger, err = logger.New(logger.Config{
		Level:  level,
		Pretty: true,
		Output: os.Stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	if cfgFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		cfgFile = filepath.Join(homeDir, ".config", "kassie", "config.json")
	}

	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		appLogger.Warn("config file not found, using defaults")
		appConfig = getDefaultConfig()
		return nil
	}

	loader := config.NewLoaderWithPath(cfgFile)
	appConfig, err = loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := appConfig.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	appLogger.With().Str("config_file", cfgFile).Logger().Info("config loaded")
	return nil
}

func getDefaultConfig() *config.Config {
	return &config.Config{
		Version: "1.0",
		Profiles: []config.Profile{
			{
				Name:     "local",
				Hosts:    []string{"127.0.0.1"},
				Port:     9042,
				Keyspace: "system",
			},
		},
		Defaults: config.DefaultConfig{
			DefaultProfile: "local",
			PageSize:       100,
			TimeoutMs:      10000,
		},
		Clients: config.ClientConfig{
			TUI: config.TUIConfig{
				Theme:   "default",
				VimMode: false,
			},
			Web: config.WebConfig{
				AutoOpenBrowser: true,
				DefaultPort:     8080,
			},
		},
	}
}

func Execute() error {
	return NewRootCmd().Execute()
}
