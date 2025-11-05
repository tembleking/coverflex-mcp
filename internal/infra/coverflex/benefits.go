package coverflex

import "log/slog"

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

	var response BenefitsResponse
	if err := c.get("https://menhir-api.coverflex.com/api/employee/benefits", &response); err != nil {
		return nil, err
	}

	return response.Benefits, nil
}
