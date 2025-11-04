package coverflex

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

const compensationURL = "https://menhir-api.coverflex.com/api/employee/compensation"

// GetCompensation fetches employee compensation from the API, handling token refresh.
func (c *Client) GetCompensation(authToken, refreshToken string) {
	slog.Info("Fetching employee compensation...")
	req, err := http.NewRequest("GET", compensationURL, nil)
	if err != nil {
		slog.Error("Error creating compensation request", "error", err)
		return
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Error fetching compensation", "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close compensation response body", "error", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var result interface{} // Use interface{} for arbitrary JSON structure
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			slog.Error("Error decoding compensation response", "error", err)
			return
		}
		slog.Info("Compensation data", "compensation", result)

	case http.StatusUnauthorized:
		slog.Info("Token expired while fetching compensation. Attempting to refresh.")
		newAuthToken, newRefreshToken := c.RefreshTokens(refreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			c.GetCompensation(newAuthToken, newRefreshToken)
		} else {
			slog.Error("Could not refresh token. Please log in again.")
			if err := c.tokenRepo.DeleteTokens(); err != nil {
				slog.Warn("failed to remove token files", "error", err)
			}
		}
	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("An HTTP error occurred while fetching compensation", "status_code", resp.StatusCode, "response", string(bodyBytes))
	}
}
