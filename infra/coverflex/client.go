package coverflex

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/tembleking/coverflex-mcp/domain"
)

// API endpoints
const (
	sessionURL    = "https://menhir-api.coverflex.com/api/employee/sessions"
	trustURL      = "https://menhir-api.coverflex.com/api/employee/sessions/trust-user-agent"
	refreshURL    = "https://menhir-api.coverflex.com/api/employee/sessions/renew"
	operationsURL = "https://menhir-api.coverflex.com/api/employee/operations"
)

// Client is the Coverflex API client.
type Client struct {
	httpClient *http.Client
	tokenRepo  domain.TokenRepository
}

// NewClient creates a new Coverflex API client.
func NewClient(tokenRepo domain.TokenRepository) *Client {
	return &Client{
		httpClient: &http.Client{},
		tokenRepo:  tokenRepo,
	}
}

// IsLoggedIn checks if the user is logged in by verifying the existence of tokens.
func (c *Client) IsLoggedIn() bool {
	_, err := c.tokenRepo.GetTokens()
	return err == nil
}

// get handles the common logic for making a GET request to the Coverflex API.
// It manages token retrieval, authorization headers, request execution, response decoding,
// and automatic token refresh on a 401 Unauthorized status.
func (c *Client) get(url string, target interface{}) error {
	tokens, err := c.tokenRepo.GetTokens()
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Initial request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", "Bearer "+tokens.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", "error", err)
		}
	}()

	// Handle token refresh and retry
	if resp.StatusCode == http.StatusUnauthorized {
		slog.Info("Token expired. Refreshing...")
		newAuthToken, _ := c.RefreshTokens(tokens.RefreshToken)
		if newAuthToken == "" {
			_ = c.tokenRepo.DeleteTokens()
			return fmt.Errorf("token refresh failed")
		}

		slog.Info("Retrying request with new token...")
		req.Header.Set("authorization", "Bearer "+newAuthToken)
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error performing retry request: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				slog.Warn("failed to close retry response body", "error", err)
			}
		}()
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d\nResponse: %s", resp.StatusCode, string(bodyBytes))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	return nil
}

// Structs for JSON payloads
type sessionRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OTP      string `json:"otp,omitempty"`
}

type otpResponse struct {
	PhoneLastDigits string `json:"phone_last_digits"`
}

type tokenResponse struct {
	Token          string `json:"token"`
	RefreshToken   string `json:"refresh_token"`
	UserAgentToken string `json:"user_agent_token"`
}

// renewTokenResponse defines the structure for the /renew endpoint.
type renewTokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type renewTokenResponse struct {
	Data renewTokenData `json:"data"`
}