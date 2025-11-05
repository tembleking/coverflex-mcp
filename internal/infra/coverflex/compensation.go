package coverflex

import "log/slog"

// Balance represents the balance of an attribution or benefit.
type Balance struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

// Attribution represents a type of compensation attribution.
type Attribution struct {
	ID      string  `json:"id"`
	Slug    string  `json:"slug"`
	Balance Balance `json:"balance"`
}

// CompensationBenefit represents a benefit within the compensation summary.
type CompensationBenefit struct {
	Slug    string  `json:"slug"`
	Balance Balance `json:"balance"`
}

// CompensationSummary is the main data structure for the compensation summary.
type CompensationSummary struct {
	Attributions []Attribution         `json:"attributions"`
	Benefits     []CompensationBenefit `json:"benefits"`
	RenewalDate  string                `json:"renewal_date"`
	Status       string                `json:"status"`
}

// CompensationResponse is the top-level structure for the compensation API response.
type CompensationResponse struct {
	Summary CompensationSummary `json:"summary"`
}

// GetCompensation fetches the employee's compensation summary from the Coverflex API.
// It automatically handles token refresh if the current token is expired.
// It returns a pointer to a CompensationSummary struct containing detailed information about compensation
// or an error if the request fails or the response cannot be decoded.
func (c *Client) GetCompensation() (*CompensationSummary, error) {
	slog.Info("Fetching employee compensation...")

	var response CompensationResponse
	if err := c.get("https://menhir-api.coverflex.com/api/employee/compensation", &response); err != nil {
		return nil, err
	}

	return &response.Summary, nil
}
