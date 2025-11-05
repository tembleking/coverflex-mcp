package coverflex

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
)

// Amount represents a monetary value and its currency.
type Amount struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

// DescriptionParam provides details for an operation's description.
type DescriptionParam struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// Operation represents a single financial operation.
type Operation struct {
	ID                string             `json:"id"`
	Amount            Amount             `json:"amount"`
	CategorySlug      string             `json:"category_slug"`
	DescriptionParams []DescriptionParam `json:"description_params"`
	DescriptionTag    string             `json:"description_tag"`
	ExecutedAt        string             `json:"executed_at"`
	IsDebit           bool               `json:"is_debit"`
	MerchantName      *string            `json:"merchant_name"`
	ProductSlug       string             `json:"product_slug"`
	Status            string             `json:"status"`
	Type              string             `json:"type"`
}

// OperationsResponse is the top-level structure for the operations API response.
type OperationsResponse struct {
	Operations struct {
		List []Operation `json:"list"`
	} `json:"operations"`
}

// OperationsFilters holds the filter parameters for GetOperations.
type OperationsFilters struct {
	Type string
}

// GetOperationsParams holds the parameters for the GetOperations method.
type GetOperationsParams struct {
	Page    int
	PerPage int
	Filters OperationsFilters
}

// GetOperationsOption defines a function that modifies GetOperationsParams.
type GetOperationsOption func(*GetOperationsParams)

// WithOperationsPage sets the page number for the operations request.
func WithOperationsPage(page int) GetOperationsOption {
	return func(params *GetOperationsParams) {
		if page > 0 {
			params.Page = page
		}
	}
}

// WithOperationsPerPage sets the number of operations to return per page.
func WithOperationsPerPage(perPage int) GetOperationsOption {
	return func(params *GetOperationsParams) {
		if perPage > 0 {
			params.PerPage = perPage
		}
	}
}

// WithOperationsFilterType sets the type filter for the operations request.
func WithOperationsFilterType(filterType string) GetOperationsOption {
	return func(params *GetOperationsParams) {
		params.Filters.Type = filterType
	}
}

// RefreshTokens handles the token refresh logic.
func (c *Client) RefreshTokens(refreshToken string) (newAuthToken, newRefreshToken string) {
	slog.Info("Attempting to refresh tokens...")

	req, err := http.NewRequest("POST", refreshURL, nil) // No body for this request
	if err != nil {
		slog.Error("Error creating refresh request", "error", err)
		return "", ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+refreshToken) // Token in header

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Error during token refresh request", "error", err)
		return "", ""
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close refresh response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated { // 200 or 201
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("Unexpected status code during token refresh", "status_code", resp.StatusCode, "response", string(bodyBytes))
		return "", ""
	}

	var renewedTokens renewTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&renewedTokens); err != nil {
		slog.Error("Error decoding refreshed token response", "error", err)
		return "", ""
	}

	newAuthToken = renewedTokens.Data.AccessToken
	newRefreshToken = renewedTokens.Data.RefreshToken

	if newAuthToken == "" || newRefreshToken == "" {
		slog.Error("Failed to retrieve new tokens from refresh response.")
		return "", ""
	}

	if err := c.tokenRepo.SaveTokens(newAuthToken, newRefreshToken); err != nil {
		slog.Error("Error saving new tokens", "error", err)
		// Continue anyway, as we have the tokens in memory
	}

	slog.Info("Tokens refreshed and saved successfully.")
	return newAuthToken, newRefreshToken
}

// GetOperations fetches financial operations from the Coverflex API.
// It supports pagination and filtering through functional options.
// It automatically handles token refresh if the current token is expired.
// It returns a slice of Operation structs or an error if the request fails.
func (c *Client) GetOperations(opts ...GetOperationsOption) ([]Operation, error) {
	slog.Info("Fetching recent operations...")

	params := &GetOperationsParams{
		Page:    1,
		PerPage: 50,
	}
	for _, opt := range opts {
		opt(params)
	}

	baseURL, err := url.Parse(operationsURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing operations URL: %w", err)
	}
	queryParams := url.Values{}
	if params.Page > 0 {
		queryParams.Add("page", strconv.Itoa(params.Page))
	}
	if params.PerPage > 0 {
		queryParams.Add("per_page", strconv.Itoa(params.PerPage))
	}
	if params.Filters.Type != "" {
		queryParams.Add(fmt.Sprintf("filters[%s]", "type"), params.Filters.Type)
	}
	baseURL.RawQuery = queryParams.Encode()

	var response OperationsResponse
	if err := c.get(baseURL.String(), &response); err != nil {
		return nil, err
	}

	return response.Operations.List, nil
}
