package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
	"github.com/tembleking/coverflex-mcp/infra/fs"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "coverflex-mcp",
	Short: "A CLI tool to interact with Coverflex services",
	Long: `coverflex-mcp is a powerful command-line interface (CLI) tool designed to streamline
your interactions with Coverflex services. It allows you to authenticate, manage tokens,
and perform various operations directly from your terminal.

Use 'coverflex-mcp [command] --help' for more information about a specific command.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
		slog.SetDefault(logger)

		tokenRepo := fs.NewTokenRepository()
		client := coverflex.NewClient(tokenRepo)

		// Default - Use existing tokens
		if client.IsLoggedIn() {
			slog.Info("Token files found. Reading tokens and fetching data.")
			if operations, err := client.GetOperations(); err != nil {
				slog.Error("Failed to get operations", "error", err)
			} else {
				slog.Info("Operations data", "operations", operations)
			}

			if benefits, err := client.GetBenefits(); err != nil {
				slog.Error("Failed to get benefits", "error", err)
			} else {
				slog.Info("Benefits data", "benefits", benefits)
			}

			if compensation, err := client.GetCompensation(); err != nil {
				slog.Error("Failed to get compensation", "error", err)
			} else {
				slog.Info("Compensation data", "compensation", compensation)
			}

			if family, err := client.GetFamily(); err != nil {
				slog.Error("Failed to get family information", "error", err)
			} else {
				slog.Info("Family data", "family", family)
			}
		} else {
			slog.Info("Token files not found. Please log in using the --user and --pass flags.")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.coverflex-mcp.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
