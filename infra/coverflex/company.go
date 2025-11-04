package coverflex

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

const companyURL = "https://menhir-api.coverflex.com/api/employee/company"

// Address represents a company address.
type Address struct {
	AddressLine1 string  `json:"address_line_1"`
	AddressLine2 *string `json:"address_line_2"`
	City         string  `json:"city"`
	Country      string  `json:"country"`
	District     string  `json:"district"`
	Type         string  `json:"type"`
	Zipcode      string  `json:"zipcode"`
}

// Market represents the market information for a company.
type Market struct {
	Languages []string `json:"languages"`
	Slug      string   `json:"slug"`
}

// Settings represents the company settings.
type Settings struct {
	CardRequestEmployeePermission string `json:"card_request_employee_permission"`
	CardRequestFormat             string `json:"card_request_format"`
	CardRequestStrategy           string `json:"card_request_strategy"`
	CardShippingStrategy          string `json:"card_shipping_strategy"`
	IncludeEmployeeNumberInReports bool  `json:"include_employee_number_in_reports"`
	KinshipDegreeProofRequired    bool  `json:"kinship_degree_proof_required"`
	Plan                          string `json:"plan"`
	SavingsEmployeeEnabled        bool  `json:"savings_employee_enabled"`
}

// TaxID represents the tax ID of a company.
type TaxID struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Company represents the company information.
type Company struct {
	ID               string    `json:"id"`
	Addresses        []Address `json:"addresses"`
	CardDisplayName  string    `json:"card_display_name"`
	LegalName        string    `json:"legal_name"`
	LogoURI          *string   `json:"logo_uri"`
	Market           Market    `json:"market"`
	Name             string    `json:"name"`
	Settings         Settings  `json:"settings"`
	TaxID            TaxID     `json:"tax_id"`
}

// CompensationConfig represents the compensation configuration for a company.
type CompensationConfig struct {
	HasSocialBenefits bool `json:"has_social_benefits"`
}

// CompanyResponse is the top-level structure for the company API response.
type CompanyResponse struct {
	Company            Company            `json:"company"`
	CompensationConfig CompensationConfig `json:"compensation_config"`
}

// GetCompany fetches employee company information from the API, handling token refresh.
func (c *Client) GetCompany() (*CompanyResponse, error) {
	slog.Info("Fetching employee company information...")

	tokens, err := c.tokenRepo.GetTokens()
	if err != nil {
		slog.Error("Not logged in. Please log in first.", "error", err)
		return nil, err
	}

	req, err := http.NewRequest("GET", companyURL, nil)
	if err != nil {
		slog.Error("Error creating company request", "error", err)
		return nil, err
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", "Bearer "+tokens.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Error fetching company information", "error", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close company response body", "error", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var companyResponse CompanyResponse
		if err := json.NewDecoder(resp.Body).Decode(&companyResponse); err != nil {
			slog.Error("Error decoding company response", "error", err)
			return nil, err
		}
		return &companyResponse, nil

	case http.StatusUnauthorized:
		slog.Info("Token expired while fetching company information. Attempting to refresh.")
		newAuthToken, _ := c.RefreshTokens(tokens.RefreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			return c.GetCompany()
		}

		err := c.tokenRepo.DeleteTokens()
		if err != nil {
			slog.Warn("failed to remove token files", "error", err)
		}
		return nil, err

	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("An HTTP error occurred while fetching company information", "status_code", resp.StatusCode, "response", string(bodyBytes))
		return nil, err
	}
}