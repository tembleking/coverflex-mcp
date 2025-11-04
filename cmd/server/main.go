package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/tembleking/coverflex-mcp/infra/coverflex"
	"github.com/tembleking/coverflex-mcp/infra/fs"
)

func main() {
	tokenRepo := fs.NewTokenRepository()
	client := coverflex.NewClient(tokenRepo)

	// --- Flag Definitions ---
	user := flag.String("user", "", "Your Coverflex email")
	pass := flag.String("pass", "", "Your Coverflex password")
	otp := flag.String("otp", "", "The OTP you received via SMS")
	forceRefresh := flag.Bool("force-refresh", false, "Force a refresh of the authentication tokens")

	flag.Parse()

	// --- Logic based on flags ---

	// Case 1: Force Refresh
	if *forceRefresh {
		fmt.Println("Force refresh option detected.")
		tokens, err := tokenRepo.GetTokens()
		if err != nil {
			log.Fatalf("Refresh token file not found. Cannot force refresh. Please log in first.")
		}
		newAuthToken, newRefreshToken := client.RefreshTokens(tokens.RefreshToken)
		if newAuthToken != "" {
			fmt.Println("\nTokens have been refreshed. Let's test the new token:")
			client.GetOperations(newAuthToken, newRefreshToken)
		} else {
			log.Fatalf("Failed to refresh tokens.")
		}
		return
	}

	// Case 2: Login with OTP
	if *user != "" && *pass != "" && *otp != "" {
		fmt.Println("User, password, and OTP provided. Attempting to log in...")
		if err := client.Login(*user, *pass, *otp); err != nil {
			log.Fatalf("Login failed: %v", err)
		}
		fmt.Println("Logged in.")
		return
	}

	// Case 3: Request OTP
	if *user != "" && *pass != "" {
		fmt.Println("User and password provided. Requesting OTP...")
		if err := client.RequestOTP(*user, *pass); err != nil {
			log.Fatalf("Failed to request OTP: %v", err)
		}
		fmt.Println("An OTP has been sent to your phone. Please re-run the command with the --otp flag.")
		return
	}

	// Case 4: Default - Use existing tokens
	tokens, err := tokenRepo.GetTokens()
	if err == nil {
		fmt.Println("Token files found. Reading tokens and fetching operations.")
		client.GetOperations(tokens.AccessToken, tokens.RefreshToken)
	} else {
		fmt.Println("Token files not found. Please log in using the --user and --pass flags.")
	}
}
