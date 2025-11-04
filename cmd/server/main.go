package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tembleking/coverflex-mcp/infra/coverflex"
)

func main() {
	client := coverflex.NewClient()

	tmpDir := os.TempDir()
	tokenPath := filepath.Join(tmpDir, "coverflex_token.txt")
	refreshTokenPath := filepath.Join(tmpDir, "coverflex_refresh_token.txt")

	forceRefresh := len(os.Args) > 1 && os.Args[1] == "--force-refresh"

	if forceRefresh {
		fmt.Println("Force refresh option detected.")
		refreshBytes, err := os.ReadFile(refreshTokenPath)
		if err != nil {
			log.Fatalf("Refresh token file not found. Cannot force refresh. Please log in first.")
		}
		refresh := strings.TrimSpace(string(refreshBytes))
		newAuthToken, newRefreshToken := client.RefreshTokens(refresh)
		if newAuthToken != "" {
			fmt.Println("\nTokens have been refreshed. Let's test the new token:")
			client.GetOperations(newAuthToken, newRefreshToken)
		} else {
			log.Fatalf("Failed to refresh tokens.")
		}
	} else {
		_, errToken := os.Stat(tokenPath)
		_, errRefreshToken := os.Stat(refreshTokenPath)

		if errToken == nil && errRefreshToken == nil {
			fmt.Println("Token files found. Reading tokens and fetching operations.")
			tokenBytes, _ := os.ReadFile(tokenPath)
			refreshBytes, _ := os.ReadFile(refreshTokenPath)
			token := strings.TrimSpace(string(tokenBytes))
			refresh := strings.TrimSpace(string(refreshBytes))
			client.GetOperations(token, refresh)
		} else {
			fmt.Println("Token files not found. Starting login process.")
			client.Login()
		}
	}
}
