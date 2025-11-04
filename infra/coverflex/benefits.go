package coverflex

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

const benefitsURL = "https://menhir-api.coverflex.com/api/employee/benefits"

// BenefitLimit defines the monetary limits for a benefit.
type BenefitLimit struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

// BenefitLimits contains monthly and yearly limits.
type BenefitLimits struct {
	Monthly *BenefitLimit `json:"monthly"`
	Yearly  *BenefitLimit `json:"yearly"`
}

// Product represents a specific product within a benefit category.
type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Type        string `json:"type"`
}

// Benefit represents a single employee benefit.
type Benefit struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Slug        string        `json:"slug"`
	Description *string       `json:"description"`
	Limits      BenefitLimits `json:"limits"`
	Products    []Product     `json:"products"`
}

// BenefitsResponse is the top-level structure for the benefits API response.
type BenefitsResponse struct {
	Benefits []Benefit `json:"benefits"`
}

// GetBenefits fetches the employee's benefits from the Coverflex API.
// It automatically handles token refresh if the current token is expired.
// It returns a slice of Benefit structs containing detailed information about each benefit
// or an error if the request fails or the response cannot be decoded.
func (c *Client) GetBenefits() ([]Benefit, error) {
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
		var benefitsResponse BenefitsResponse
		if err := json.NewDecoder(resp.Body).Decode(&benefitsResponse); err != nil {
			slog.Error("Error decoding benefits response", "error", err)
			return nil, err
		}
		return benefitsResponse.Benefits, nil

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
