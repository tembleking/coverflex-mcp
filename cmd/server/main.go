package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tembleking/coverflex-mcp/infra/coverflex"
	"github.com/tembleking/coverflex-mcp/infra/fs"
)

func main() {
	tokenRepo := fs.NewTokenRepository()
	client := coverflex.NewClient(tokenRepo)

	forceRefresh := len(os.Args) > 1 && os.Args[1] == "--force-refresh"

	if forceRefresh {
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
	} else {
		tokens, err := tokenRepo.GetTokens()
		if err == nil {
			fmt.Println("Token files found. Reading tokens and fetching operations.")
			client.GetOperations(tokens.AccessToken, tokens.RefreshToken)
		} else {
			fmt.Println("Token files not found. Starting login process.")
			client.Login()
		}
	}
}
