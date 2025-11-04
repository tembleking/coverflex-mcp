package coverflex

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

const cardsURL = "https://menhir-api.coverflex.com/api/employee/cards"

// Card represents a single employee card.
type Card struct {
	ID                string  `json:"id"`
	ActivatedAt       *string `json:"activated_at"`
	ExpirationDate    string  `json:"expiration_date"`
	Format            string  `json:"format"`
	HolderCompanyName string  `json:"holder_company_name"`
	HolderName        string  `json:"holder_name"`
	IsExpiring        bool    `json:"is_expiring"`
	IsPlasticRequested bool   `json:"is_plastic_requested"`
	Network           string  `json:"network"`
	OwnerID           string  `json:"owner_id"`
	PANLastDigits     string  `json:"pan_last_digits"`
	ProviderID        string  `json:"provider_id"`
	Status            string  `json:"status"`
	Version           string  `json:"version"`
	// Add other fields as necessary, using *string for nullable fields
}

// CardsResponse is the top-level structure for the cards API response.
type CardsResponse struct {
	Cards []Card `json:"cards"`
}

// GetCards fetches the employee's cards from the Coverflex API.
// It automatically handles token refresh if the current token is expired.
// It returns a slice of Card structs containing detailed information about each card
// or an error if the request fails or the response cannot be decoded.
func (c *Client) GetCards() ([]Card, error) {
	slog.Info("Fetching employee cards information...")

	tokens, err := c.tokenRepo.GetTokens()
	if err != nil {
		slog.Error("Not logged in. Please log in first.", "error", err)
		return nil, err
	}

	req, err := http.NewRequest("GET", cardsURL, nil)
	if err != nil {
		slog.Error("Error creating cards request", "error", err)
		return nil, err
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", "Bearer "+tokens.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Error fetching cards information", "error", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close cards response body", "error", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var cardsResponse CardsResponse
		if err := json.NewDecoder(resp.Body).Decode(&cardsResponse); err != nil {
			slog.Error("Error decoding cards response", "error", err)
			return nil, err
		}
		return cardsResponse.Cards, nil

	case http.StatusUnauthorized:
		slog.Info("Token expired while fetching cards information. Attempting to refresh.")
		newAuthToken, _ := c.RefreshTokens(tokens.RefreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			return c.GetCards()
		}

		err := c.tokenRepo.DeleteTokens()
		if err != nil {
			slog.Warn("failed to remove token files", "error", err)
		}
		return nil, err

	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("An HTTP error occurred while fetching cards information", "status_code", resp.StatusCode, "response", string(bodyBytes))
		return nil, err
	}
}