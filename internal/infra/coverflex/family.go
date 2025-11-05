package coverflex

import "log/slog"

// FamilyMember represents a single member of the employee's family.
type FamilyMember struct {
	ID           string `json:"id"`
	FullName     string `json:"full_name"`
	ShortName    string `json:"short_name"`
	BirthDate    string `json:"birth_date"`
	RelationType string `json:"relation_type"`
	Gender       string `json:"gender"`
}

// FamilyResponse is the top-level structure for the family API response.
type FamilyResponse struct {
	Members []FamilyMember `json:"members"`
}

// GetFamily fetches the employee's family members from the Coverflex API.
// It automatically handles token refresh if the current token is expired.
// It returns a slice of FamilyMember structs containing detailed information about each family member
// or an error if the request fails or the response cannot be decoded.
func (c *Client) GetFamily() ([]FamilyMember, error) {
	slog.Info("Fetching employee family information...")

	var response FamilyResponse
	if err := c.get("https://menhir-api.coverflex.com/api/employee/family", &response); err != nil {
		return nil, err
	}

	return response.Members, nil
}
