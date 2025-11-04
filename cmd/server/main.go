package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/tembleking/coverflex-mcp/infra/coverflex"
	"github.com/tembleking/coverflex-mcp/infra/fs"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tokenRepo := fs.NewTokenRepository(logger)
	client := coverflex.NewClient(tokenRepo, logger)

	// --- Flag Definitions ---
	user := flag.String("user", "", "Your Coverflex email")
	pass := flag.String("pass", "", "Your Coverflex password")
	otp := flag.String("otp", "", "The OTP you received via SMS")
	forceRefresh := flag.Bool("force-refresh", false, "Force a refresh of the authentication tokens")

	flag.Parse()

	// --- Logic based on flags ---

	// Case 1: Force Refresh
	if *forceRefresh {
		logger.Info("Force refresh option detected.")
		tokens, err := tokenRepo.GetTokens()
		if err != nil {
			logger.Error("Refresh token file not found. Cannot force refresh. Please log in first.")
			os.Exit(1)
		}
		newAuthToken, newRefreshToken := client.RefreshTokens(tokens.RefreshToken)
		if newAuthToken != "" {
			logger.Info("\nTokens have been refreshed. Let's test the new token:")
			client.GetOperations(newAuthToken, newRefreshToken)
		} else {
			logger.Error("Failed to refresh tokens.")
			os.Exit(1)
		}
		return
	}

	// Case 2: Login with OTP
	if *user != "" && *pass != "" && *otp != "" {
		logger.Info("User, password, and OTP provided. Attempting to log in...")
		if err := client.Login(*user, *pass, *otp); err != nil {
			logger.Error("Login failed", "error", err)
			os.Exit(1)
		}
		logger.Info("Logged in.")
		return
	}

	// Case 3: Request OTP
	if *user != "" && *pass != "" {
		logger.Info("User and password provided. Requesting OTP...")
		if err := client.RequestOTP(*user, *pass); err != nil {
			logger.Error("Failed to request OTP", "error", err)
			os.Exit(1)
		}
		logger.Info("An OTP has been sent to your phone. Please re-run the command with the --otp flag.")
		return
	}

	// Case 4: Default - Use existing tokens
	tokens, err := tokenRepo.GetTokens()
	if err == nil {
		logger.Info("Token files found. Reading tokens and fetching operations.")
		client.GetOperations(tokens.AccessToken, tokens.RefreshToken)
	} else {
		logger.Info("Token files not found. Please log in using the --user and --pass flags.")
	}
}
