package coverflex

import "log/slog"

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
	CardRequestEmployeePermission  string `json:"card_request_employee_permission"`
	CardRequestFormat              string `json:"card_request_format"`
	CardRequestStrategy            string `json:"card_request_strategy"`
	CardShippingStrategy           string `json:"card_shipping_strategy"`
	IncludeEmployeeNumberInReports bool   `json:"include_employee_number_in_reports"`
	KinshipDegreeProofRequired     bool   `json:"kinship_degree_proof_required"`
	Plan                           string `json:"plan"`
	SavingsEmployeeEnabled         bool   `json:"savings_employee_enabled"`
}

// TaxID represents the tax ID of a company.
type TaxID struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Company represents the company information.
type Company struct {
	ID              string    `json:"id,omitempty"`
	Addresses       []Address `json:"addresses,omitempty"`
	CardDisplayName string    `json:"card_display_name,omitempty"`
	LegalName       string    `json:"legal_name,omitempty"`
	LogoURI         *string   `json:"logo_uri,omitempty"`
	Market          Market    `json:"market,omitempty"`
	Name            string    `json:"name,omitempty"`
	Settings        Settings  `json:"settings,omitempty"`
	TaxID           TaxID     `json:"tax_id,omitempty"`
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

// GetCompany fetches the employee's company information from the Coverflex API.
// It automatically handles token refresh if the current token is expired.
// It returns a pointer to a CompanyResponse struct containing detailed information about the company
// or an error if the request fails or the response cannot be decoded.
func (c *Client) GetCompany() (*CompanyResponse, error) {
	slog.Info("Fetching employee company information...")

	var response CompanyResponse
	if err := c.get("https://menhir-api.coverflex.com/api/employee/company", &response); err != nil {
		return nil, err
	}

	return &response, nil
}
