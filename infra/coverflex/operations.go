package coverflex

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// RefreshTokens handles the token refresh logic.
func (c *Client) RefreshTokens(refreshToken string) (newAuthToken, newRefreshToken string) {
	slog.Info("Attempting to refresh tokens...")

	req, err := http.NewRequest("POST", refreshURL, nil) // No body for this request
	if err != nil {
		slog.Error("Error creating refresh request", "error", err)
		return "", ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+refreshToken) // Token in header

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Error during token refresh request", "error", err)
		return "", ""
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close refresh response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated { // 200 or 201
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("Unexpected status code during token refresh", "status_code", resp.StatusCode, "response", string(bodyBytes))
		return "", ""
	}

	var renewedTokens renewTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&renewedTokens); err != nil {
		slog.Error("Error decoding refreshed token response", "error", err)
		return "", ""
	}

	newAuthToken = renewedTokens.Data.AccessToken
	newRefreshToken = renewedTokens.Data.RefreshToken

	if newAuthToken == "" || newRefreshToken == "" {
		slog.Error("Failed to retrieve new tokens from refresh response.")
		return "", ""
	}

	if err := c.tokenRepo.SaveTokens(newAuthToken, newRefreshToken); err != nil {
		slog.Error("Error saving new tokens", "error", err)
		// Continue anyway, as we have the tokens in memory
	}

	slog.Info("Tokens refreshed and saved successfully.")
	return newAuthToken, newRefreshToken
}

// GetOperations fetches the 5 most recent operations from the API, handling token refresh.
func (c *Client) GetOperations(authToken, refreshToken string) {
	slog.Info("\nFetching recent operations...")
	req, err := http.NewRequest("GET", operationsURL, nil)
	if err != nil {
		slog.Error("Error creating operations request", "error", err)
		return
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "es-ES,es;q=0.9")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Error fetching operations", "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close operations response body", "error", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var result interface{} // Use interface{} for arbitrary JSON structure
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			slog.Error("Error decoding operations response", "error", err)
			return
		}
		slog.Info("Operations data", "operations", result)

	case http.StatusUnauthorized:
		slog.Info("Token expired.")
		newAuthToken, newRefreshToken := c.RefreshTokens(refreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			c.GetOperations(newAuthToken, newRefreshToken)
		} else {
			slog.Error("Could not refresh token. Please log in again.")
			if err := c.tokenRepo.DeleteTokens(); err != nil {
				slog.Warn("failed to remove token files", "error", err)
			}
		}
	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("An HTTP error occurred", "status_code", resp.StatusCode, "response", string(bodyBytes))
	}
}
