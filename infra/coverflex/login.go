package coverflex

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Login handles the user login process.
func (c *Client) Login() {
	reader := bufio.NewReader(os.Stdin)

	email, password := c.getCredentials(reader)

	if err := c.requestOTP(email, password); err != nil {
		log.Fatalf("OTP request failed: %v", err)
	}

	authToken, refreshToken, err := c.submitOTP(reader, email, password)
	if err != nil {
		log.Fatalf("OTP submission failed: %v", err)
	}

	authToken, refreshToken = c.trustDevice(authToken, refreshToken)

	if err := c.tokenRepo.SaveTokens(authToken, refreshToken); err != nil {
		log.Fatalf("Error saving tokens: %v", err)
	}

	if authToken != "" && refreshToken != "" {
		c.GetOperations(authToken, refreshToken)
	}
}

func (c *Client) getCredentials(reader *bufio.Reader) (string, string) {
	fmt.Print("Enter your email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Enter your password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	return email, password
}

func (c *Client) requestOTP(email, password string) error {
	fmt.Println("Requesting OTP...")
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
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusAccepted { // 202
		var otpResp otpResponse
		if err := json.NewDecoder(resp.Body).Decode(&otpResp); err != nil {
			return fmt.Errorf("error decoding OTP response: %w", err)
		}
		fmt.Printf("OTP sent to phone ending in ...%s\n", otpResp.PhoneLastDigits)
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unexpected status code during OTP request: %d\nResponse: %s", resp.StatusCode, string(bodyBytes))
}

func (c *Client) submitOTP(reader *bufio.Reader, email, password string) (string, string, error) {
	fmt.Print("Enter the OTP you received: ")
	otp, _ := reader.ReadString('\n')
	otp = strings.TrimSpace(otp)

	fmt.Println("Submitting OTP...")
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
			log.Printf("Warning: failed to close response body: %v", err)
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

	fmt.Println("Successfully authenticated.")
	return tokens.Token, tokens.RefreshToken, nil
}

func (c *Client) trustDevice(authToken, refreshToken string) (string, string) {
	fmt.Println("Trusting this device...")
	req, err := http.NewRequest("POST", trustURL, nil)
	if err != nil {
		log.Printf("Warning: Error creating trust request: %v", err)
		return authToken, refreshToken // Return original tokens on error
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Warning: Error trusting device: %v", err)
		return authToken, refreshToken // Return original tokens on error
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusCreated { // 201
		fmt.Println("Device trusted successfully.")
		var newTokens tokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&newTokens); err == nil {
			if newTokens.UserAgentToken != "" {
				fmt.Println("Received user agent token for long-term session.")
			}
			// Return new tokens
			return newTokens.Token, newTokens.RefreshToken
		}
	}

	// If not successful, return original tokens
	return authToken, refreshToken
}
