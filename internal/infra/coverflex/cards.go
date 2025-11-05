package coverflex

import "log/slog"

// Card represents a single employee card.
type Card struct {
	ID                 string  `json:"id"`
	ActivatedAt        *string `json:"activated_at"`
	ExpirationDate     string  `json:"expiration_date"`
	Format             string  `json:"format"`
	HolderCompanyName  string  `json:"holder_company_name"`
	HolderName         string  `json:"holder_name"`
	IsExpiring         bool    `json:"is_expiring"`
	IsPlasticRequested bool    `json:"is_plastic_requested"`
	Network            string  `json:"network"`
	OwnerID            string  `json:"owner_id"`
	PANLastDigits      string  `json:"pan_last_digits"`
	ProviderID         string  `json:"provider_id"`
	Status             string  `json:"status"`
	Version            string  `json:"version"`
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

	var response CardsResponse
	if err := c.get("https://menhir-api.coverflex.com/api/employee/cards", &response); err != nil {
		return nil, err
	}

	return response.Cards, nil
}
