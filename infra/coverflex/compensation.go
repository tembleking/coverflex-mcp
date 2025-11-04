package coverflex

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

const compensationURL = "https://menhir-api.coverflex.com/api/employee/compensation"

// GetCompensation fetches employee compensation from the API, handling token refresh.
func (c *Client) GetCompensation() (map[string]interface{}, error) {
	slog.Info("Fetching employee compensation...")

	tokens, err := c.tokenRepo.GetTokens()
	if err != nil {
		slog.Error("Not logged in. Please log in first.", "error", err)
		return nil, err
	}

	req, err := http.NewRequest("GET", compensationURL, nil)
	if err != nil {
		slog.Error("Error creating compensation request", "error", err)
		return nil, err
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", "Bearer "+tokens.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Error fetching compensation", "error", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close compensation response body", "error", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var data map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			slog.Error("Error decoding compensation response", "error", err)
			return nil, err
		}
		return data, nil

	case http.StatusUnauthorized:
		slog.Info("Token expired while fetching compensation. Attempting to refresh.")
		newAuthToken, _ := c.RefreshTokens(tokens.RefreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			return c.GetCompensation()
		}

		err := c.tokenRepo.DeleteTokens()
		if err != nil {
			slog.Warn("failed to remove token files", "error", err)
		}
		return nil, err

	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("An HTTP error occurred while fetching compensation", "status_code", resp.StatusCode, "response", string(bodyBytes))
		return nil, err
	}
}
