package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
	"github.com/tembleking/coverflex-mcp/infra/fs"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Coverflex",
	Long:  `This command allows you to log in to Coverflex using your email, password, and an optional OTP. Use --force-refresh to renew tokens.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
		slog.SetDefault(logger)

		tokenRepo := fs.NewTokenRepository()
		client := coverflex.NewClient(tokenRepo)

		user, _ := cmd.Flags().GetString("user")
		pass, _ := cmd.Flags().GetString("pass")
		otp, _ := cmd.Flags().GetString("otp")
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

		if user != "" && pass != "" && otp != "" {
			slog.Info("User, password, and OTP provided. Attempting to log in...")
			if err := client.Login(user, pass, otp); err != nil {
				slog.Error("Login failed", "error", err)
				os.Exit(1)
			}
			slog.Info("Logged in.")
			return
		}

		if user != "" && pass != "" {
			slog.Info("User and password provided. Requesting OTP...")
			if err := client.RequestOTP(user, pass); err != nil {
				slog.Error("Failed to request OTP", "error", err)
				os.Exit(1)
			}
			slog.Info("An OTP has been sent to your phone. Please re-run the command with the --otp flag.")
			return
		}

		slog.Error("Please provide your Coverflex email and password using --user and --pass flags. If you have received an OTP, also provide it with the --otp flag.")
		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().String("user", "", "Your Coverflex email")
	loginCmd.Flags().String("pass", "", "Your Coverflex password")
	loginCmd.Flags().StringP("otp", "o", "", "The OTP you received via SMS")
	loginCmd.Flags().Bool("force-refresh", false, "Force a refresh of the authentication tokens")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
