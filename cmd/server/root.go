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
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
		slog.SetDefault(logger)

		tokenRepo := fs.NewTokenRepository()
		client := coverflex.NewClient(tokenRepo)

		forceRefresh, _ := cmd.Flags().GetBool("force-refresh")

		// Case 1: Force Refresh
		if forceRefresh {
			slog.Info("Force refresh option detected.")
			tokens, err := tokenRepo.GetTokens()
			if err != nil {
				slog.Error("Refresh token file not found. Cannot force refresh. Please log in first.")
				os.Exit(1)
			}
			newAuthToken, newRefreshToken := client.RefreshTokens(tokens.RefreshToken)
			if newAuthToken != "" {
				slog.Info("\nTokens have been refreshed. Let's test the new token:")
				client.GetOperations(newAuthToken, newRefreshToken)
			} else {
				slog.Error("Failed to refresh tokens.")
				os.Exit(1)
			}
			return
		}

		// Case 4: Default - Use existing tokens
		tokens, err := tokenRepo.GetTokens()
		if err == nil {
			slog.Info("Token files found. Reading tokens and fetching operations.")
			client.GetOperations(tokens.AccessToken, tokens.RefreshToken)
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

	rootCmd.Flags().Bool("force-refresh", false, "Force a refresh of the authentication tokens")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
