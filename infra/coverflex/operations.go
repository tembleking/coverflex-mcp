package coverflex

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// Amount represents a monetary value and its currency.
type Amount struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

// DescriptionParam provides details for an operation's description.
type DescriptionParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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

// GetOperations fetches the 5 most recent operations from the API, handling token refresh.
func (c *Client) GetOperations() ([]Operation, error) {
	slog.Info("Fetching recent operations...")

	tokens, err := c.tokenRepo.GetTokens()
	if err != nil {
		slog.Error("Not logged in. Please log in first.", "error", err)
		return nil, err
	}

	req, err := http.NewRequest("GET", operationsURL, nil)
	if err != nil {
		slog.Error("Error creating operations request", "error", err)
		return nil, err
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "es-ES,es;q=0.9")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Error fetching operations", "error", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close operations response body", "error", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var opsResponse OperationsResponse
		if err := json.NewDecoder(resp.Body).Decode(&opsResponse); err != nil {
			slog.Error("Error decoding operations response", "error", err)
			return nil, err
		}
		return opsResponse.Operations.List, nil

	case http.StatusUnauthorized:
		slog.Info("Token expired.")
		newAuthToken, _ := c.RefreshTokens(tokens.RefreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			return c.GetOperations()
		}

		err := c.tokenRepo.DeleteTokens()
		if err != nil {
			slog.Warn("failed to remove token files", "error", err)
		}
		return nil, err

	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("An HTTP error occurred", "status_code", resp.StatusCode, "response", string(bodyBytes))
		return nil, err
	}
}