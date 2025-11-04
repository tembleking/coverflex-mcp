package coverflex

import (
	"encoding/json"
	"io"
	"net/http"
)

// RefreshTokens handles the token refresh logic.
func (c *Client) RefreshTokens(refreshToken string) (newAuthToken, newRefreshToken string) {
	c.logger.Info("Attempting to refresh tokens...")

	req, err := http.NewRequest("POST", refreshURL, nil) // No body for this request
	if err != nil {
		c.logger.Error("Error creating refresh request", "error", err)
		return "", ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+refreshToken) // Token in header

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Error during token refresh request", "error", err)
		return "", ""
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warn("failed to close refresh response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated { // 200 or 201
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logger.Error("Unexpected status code during token refresh", "status_code", resp.StatusCode, "response", string(bodyBytes))
		return "", ""
	}

	var renewedTokens renewTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&renewedTokens); err != nil {
		c.logger.Error("Error decoding refreshed token response", "error", err)
		return "", ""
	}

	newAuthToken = renewedTokens.Data.AccessToken
	newRefreshToken = renewedTokens.Data.RefreshToken

	if newAuthToken == "" || newRefreshToken == "" {
		c.logger.Error("Failed to retrieve new tokens from refresh response.")
		return "", ""
	}

	if err := c.tokenRepo.SaveTokens(newAuthToken, newRefreshToken); err != nil {
		c.logger.Error("Error saving new tokens", "error", err)
		// Continue anyway, as we have the tokens in memory
	}

	c.logger.Info("Tokens refreshed and saved successfully.")
	return newAuthToken, newRefreshToken
}

// GetOperations fetches the 5 most recent operations from the API, handling token refresh.
func (c *Client) GetOperations(authToken, refreshToken string) {
	c.logger.Info("\nFetching recent operations...")
	req, err := http.NewRequest("GET", operationsURL, nil)
	if err != nil {
		c.logger.Error("Error creating operations request", "error", err)
		return
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "es-ES,es;q=0.9")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Error fetching operations", "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warn("failed to close operations response body", "error", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var result interface{} // Use interface{} for arbitrary JSON structure
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.logger.Error("Error decoding operations response", "error", err)
			return
		}
		prettyJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			c.logger.Error("Error formatting JSON", "error", err)
			return
		}
		c.logger.Info("Operations data:", "operations", string(prettyJSON))

	case http.StatusUnauthorized:
		c.logger.Info("Token expired.")
		newAuthToken, newRefreshToken := c.RefreshTokens(refreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			c.GetOperations(newAuthToken, newRefreshToken)
		} else {
			c.logger.Error("Could not refresh token. Please log in again.")
			if err := c.tokenRepo.DeleteTokens(); err != nil {
				c.logger.Warn("failed to remove token files", "error", err)
			}
		}
	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logger.Error("An HTTP error occurred", "status_code", resp.StatusCode, "response", string(bodyBytes))
	}
}
