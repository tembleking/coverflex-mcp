package coverflex

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

const benefitsURL = "https://menhir-api.coverflex.com/api/employee/benefits"

// GetBenefits fetches employee benefits from the API, handling token refresh.
func (c *Client) GetBenefits() {
	slog.Info("Fetching employee benefits...")

	tokens, err := c.tokenRepo.GetTokens()
	if err != nil {
		slog.Error("Not logged in. Please log in first.", "error", err)
		return
	}

	req, err := http.NewRequest("GET", benefitsURL, nil)
	if err != nil {
		slog.Error("Error creating benefits request", "error", err)
		return
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", "Bearer "+tokens.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Error fetching benefits", "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close benefits response body", "error", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var result interface{} // Use interface{} for arbitrary JSON structure
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			slog.Error("Error decoding benefits response", "error", err)
			return
		}
		slog.Info("Benefits data", "benefits", result)

	case http.StatusUnauthorized:
		slog.Info("Token expired while fetching benefits. Attempting to refresh.")
		newAuthToken, _ := c.RefreshTokens(tokens.RefreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			c.GetBenefits()
		} else {
			slog.Error("Could not refresh token. Please log in again.")
			if err := c.tokenRepo.DeleteTokens(); err != nil {
				slog.Warn("failed to remove token files", "error", err)
			}
		}
	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("An HTTP error occurred while fetching benefits", "status_code", resp.StatusCode, "response", string(bodyBytes))
	}
}
