package coverflex

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

const familyURL = "https://menhir-api.coverflex.com/api/employee/family"

// FamilyMember represents a single member of the employee's family.
type FamilyMember struct {
	ID         string  `json:"id"`
	FullName   string  `json:"full_name"`	
	ShortName  string  `json:"short_name"`
	BirthDate  string  `json:"birth_date"`
	RelationType string `json:"relation_type"`
	Gender     string  `json:"gender"`
	// Add other fields as necessary, using *string for nullable fields
}

// FamilyResponse is the top-level structure for the family API response.
type FamilyResponse struct {
	Members []FamilyMember `json:"members"`
}

// GetFamily fetches employee family information from the API, handling token refresh.
func (c *Client) GetFamily() ([]FamilyMember, error) {
	slog.Info("Fetching employee family information...")

	tokens, err := c.tokenRepo.GetTokens()
	if err != nil {
		slog.Error("Not logged in. Please log in first.", "error", err)
		return nil, err
	}

	req, err := http.NewRequest("GET", familyURL, nil)
	if err != nil {
		slog.Error("Error creating family request", "error", err)
		return nil, err
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", "Bearer "+tokens.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Error fetching family information", "error", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close family response body", "error", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var familyResponse FamilyResponse
		if err := json.NewDecoder(resp.Body).Decode(&familyResponse); err != nil {
			slog.Error("Error decoding family response", "error", err)
			return nil, err
		}
		return familyResponse.Members, nil

	case http.StatusUnauthorized:
		slog.Info("Token expired while fetching family information. Attempting to refresh.")
		newAuthToken, _ := c.RefreshTokens(tokens.RefreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			return c.GetFamily()
		}

		err := c.tokenRepo.DeleteTokens()
		if err != nil {
			slog.Warn("failed to remove token files", "error", err)
		}
		return nil, err

	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("An HTTP error occurred while fetching family information", "status_code", resp.StatusCode, "response", string(bodyBytes))
		return nil, err
	}
}