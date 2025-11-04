package coverflex

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// RefreshTokens handles the token refresh logic.
func (c *Client) RefreshTokens(refreshToken string) (newAuthToken, newRefreshToken string) {
	fmt.Println("Attempting to refresh tokens...")

	req, err := http.NewRequest("POST", refreshURL, nil) // No body for this request
	if err != nil {
		log.Printf("Error creating refresh request: %v", err)
		return "", ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+refreshToken) // Token in header

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Error during token refresh request: %v", err)
		return "", ""
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close refresh response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated { // 200 or 201
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Unexpected status code during token refresh: %d\nResponse: %s", resp.StatusCode, string(bodyBytes))
		return "", ""
	}

	var renewedTokens renewTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&renewedTokens); err != nil {
		log.Printf("Error decoding refreshed token response: %v", err)
		return "", ""
	}

	newAuthToken = renewedTokens.Data.AccessToken
	newRefreshToken = renewedTokens.Data.RefreshToken

	if newAuthToken == "" || newRefreshToken == "" {
		log.Println("Failed to retrieve new tokens from refresh response.")
		return "", ""
	}

	// Save new tokens
	tmpDir := os.TempDir()
	tokenPath := filepath.Join(tmpDir, "coverflex_token.txt")
	refreshTokenPath := filepath.Join(tmpDir, "coverflex_refresh_token.txt")

	if err := os.WriteFile(tokenPath, []byte(newAuthToken), 0600); err != nil {
		log.Printf("Error saving new auth token: %v", err)
		// Continue anyway, as we have the tokens in memory
	}
	if err := os.WriteFile(refreshTokenPath, []byte(newRefreshToken), 0600); err != nil {
		log.Printf("Error saving new refresh token: %v", err)
	}

	fmt.Println("Tokens refreshed and saved successfully.")
	return newAuthToken, newRefreshToken
}

// GetOperations fetches the 5 most recent operations from the API, handling token refresh.
func (c *Client) GetOperations(authToken, refreshToken string) {
	fmt.Println("\nFetching recent operations...")
	req, err := http.NewRequest("GET", operationsURL, nil)
	if err != nil {
		log.Printf("Error creating operations request: %v", err)
		return
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "es-ES,es;q=0.9")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Error fetching operations: %v", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close operations response body: %v", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var result interface{} // Use interface{} for arbitrary JSON structure
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("Error decoding operations response: %v", err)
			return
		}
		prettyJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Printf("Error formatting JSON: %v", err)
			return
		}
		fmt.Println("Operations data:")
		fmt.Println(string(prettyJSON))

	case http.StatusUnauthorized:
		fmt.Println("Token expired.")
		newAuthToken, newRefreshToken := c.RefreshTokens(refreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			c.GetOperations(newAuthToken, newRefreshToken)
		} else {
			fmt.Println("Could not refresh token. Please log in again.")
			// Optionally, delete the expired token files
			tmpDir := os.TempDir()
			if err := os.Remove(filepath.Join(tmpDir, "coverflex_token.txt")); err != nil {
				log.Printf("Warning: failed to remove token file: %v", err)
			}
			if err := os.Remove(filepath.Join(tmpDir, "coverflex_refresh_token.txt")); err != nil {
				log.Printf("Warning: failed to remove refresh token file: %v", err)
			}
		}
	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("An HTTP error occurred: %d\nResponse: %s", resp.StatusCode, string(bodyBytes))
	}
}
