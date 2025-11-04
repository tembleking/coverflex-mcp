package coverflex

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

const benefitsURL = "https://menhir-api.coverflex.com/api/employee/benefits"

// GetBenefits fetches employee benefits from the API, handling token refresh.
func (c *Client) GetBenefits() ([]map[string]interface{}, error) {
	slog.Info("Fetching employee benefits...")

	tokens, err := c.tokenRepo.GetTokens()
	if err != nil {
		slog.Error("Not logged in. Please log in first.", "error", err)
		return nil, err
	}

	req, err := http.NewRequest("GET", benefitsURL, nil)
	if err != nil {
		slog.Error("Error creating benefits request", "error", err)
		return nil, err
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", "Bearer "+tokens.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Error fetching benefits", "error", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close benefits response body", "error", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var response struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			slog.Error("Error decoding benefits response", "error", err)
			return nil, err
		}
		return response.Data, nil

	case http.StatusUnauthorized:
		slog.Info("Token expired while fetching benefits. Attempting to refresh.")
		newAuthToken, _ := c.RefreshTokens(tokens.RefreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			return c.GetBenefits()
		}

		err := c.tokenRepo.DeleteTokens()
		if err != nil {
			slog.Warn("failed to remove token files", "error", err)
		}
		return nil, err

	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("An HTTP error occurred while fetching benefits", "status_code", resp.StatusCode, "response", string(bodyBytes))
		return nil, err
	}
}
