package coverflex

import (
	"net/http"

	"github.com/tembleking/coverflex-mcp/domain"
)

// API endpoints
const (
	sessionURL    = "https://menhir-api.coverflex.com/api/employee/sessions"
	trustURL      = "https://menhir-api.coverflex.com/api/employee/sessions/trust-user-agent"
	refreshURL    = "https://menhir-api.coverflex.com/api/employee/sessions/renew"
	operationsURL = "https://menhir-api.coverflex.com/api/employee/operations?per_page=5"
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
