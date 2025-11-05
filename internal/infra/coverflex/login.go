package coverflex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// RequestOTP initiates the login process by requesting an OTP.
func (c *Client) RequestOTP(email, password string) error {
	payload := sessionRequest{Email: email, Password: password}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error creating JSON payload: %w", err)
	}

	req, err := http.NewRequest("POST", sessionURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error during OTP request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusAccepted { // 202
		var otpResp otpResponse
		if err := json.NewDecoder(resp.Body).Decode(&otpResp); err != nil {
			return fmt.Errorf("error decoding OTP response: %w", err)
		}
		slog.Info("OTP sent to phone", "phone_last_digits", otpResp.PhoneLastDigits)
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unexpected status code during OTP request: %d\nResponse: %s", resp.StatusCode, string(bodyBytes))
}

// Login completes the authentication process using the provided OTP.
func (c *Client) Login(email, password, otp string) error {
	authToken, refreshToken, err := c.submitOTP(email, password, otp)
	if err != nil {
		return err
	}

	authToken, refreshToken = c.trustDevice(authToken, refreshToken)

	if err := c.tokenRepo.SaveTokens(authToken, refreshToken); err != nil {
		return fmt.Errorf("error saving tokens: %w", err)
	}

	return nil
}

func (c *Client) submitOTP(email, password, otp string) (string, string, error) {
	slog.Info("Submitting OTP...")
	payload := sessionRequest{Email: email, Password: password, OTP: otp}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", "", fmt.Errorf("error creating OTP payload: %w", err)
	}

	req, err := http.NewRequest("POST", sessionURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", "", fmt.Errorf("error creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error during token request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated { // 201
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("unexpected status code during token request: %d\nResponse: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokens tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return "", "", fmt.Errorf("error decoding token response: %w", err)
	}

	if tokens.Token == "" {
		return "", "", fmt.Errorf("failed to retrieve auth token")
	}

	slog.Info("Successfully authenticated.")
	return tokens.Token, tokens.RefreshToken, nil
}

func (c *Client) trustDevice(authToken, refreshToken string) (string, string) {
	slog.Info("Trusting this device...")
	req, err := http.NewRequest("POST", trustURL, nil)
	if err != nil {
		slog.Warn("Error creating trust request", "error", err)
		return authToken, refreshToken
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Warn("Error trusting device", "error", err)
		return authToken, refreshToken
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusCreated { // 201
		slog.Info("Device trusted successfully.")
		var newTokens tokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&newTokens); err == nil {
			if newTokens.UserAgentToken != "" {
				slog.Info("Received user agent token for long-term session.")
			}
			return newTokens.Token, newTokens.RefreshToken
		}
	}

	return authToken, refreshToken
}
