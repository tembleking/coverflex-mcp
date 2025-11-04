package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

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

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// API endpoints
const (
	sessionURL    = "https://menhir-api.coverflex.com/api/employee/sessions"
	trustURL      = "https://menhir-api.coverflex.com/api/employee/sessions/trust-user-agent"
	refreshURL    = "https://menhir-api.coverflex.com/api/employee/sessions/refresh"
	operationsURL = "https://menhir-api.coverflex.com/api/employee/operations?per_page=5"
)

// refreshTokens handles the token refresh logic.
func refreshTokens(refreshToken string) (newAuthToken, newRefreshToken string) {
	fmt.Println("Attempting to refresh tokens...")
	payload := refreshRequest{RefreshToken: refreshToken}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error creating refresh token payload: %v", err)
		return "", ""
	}

	req, err := http.NewRequest("POST", refreshURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating refresh request: %v", err)
		return "", ""
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error during token refresh request: %v", err)
		return "", ""
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close refresh response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated { // 201
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Unexpected status code during token refresh: %d\nResponse: %s", resp.StatusCode, string(bodyBytes))
		return "", ""
	}

	var tokens tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		log.Printf("Error decoding refreshed token response: %v", err)
		return "", ""
	}

	fmt.Println("Tokens refreshed successfully.")
	return tokens.Token, tokens.RefreshToken
}

// getOperations fetches the 5 most recent operations from the API, handling token refresh.
func getOperations(authToken, refreshToken string) {
	fmt.Println("\nFetching recent operations...")
	req, err := http.NewRequest("GET", operationsURL, nil)
	if err != nil {
		log.Printf("Error creating operations request: %v", err)
		return
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "es-ES,es;q=0.9")
	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching operations: %v", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close operations response body: %v", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var result interface{} // Use interface{} for arbitrary JSON structure
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("Error decoding operations response: %v", err)
			return
		}
		prettyJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Printf("Error formatting JSON: %v", err)
			return
		}
		fmt.Println("Operations data:")
		fmt.Println(string(prettyJSON))

	case http.StatusUnauthorized:
		fmt.Println("Token expired.")
		newAuthToken, newRefreshToken := refreshTokens(refreshToken)
		if newAuthToken != "" {
			// Retry the request with the new token
			getOperations(newAuthToken, newRefreshToken)
		} else {
			fmt.Println("Could not refresh token. Please log in again.")
			// Optionally, delete the expired token files
			tmpDir := os.TempDir()
			if err := os.Remove(filepath.Join(tmpDir, "coverflex_token.txt")); err != nil {
				log.Printf("Warning: failed to remove token file: %v", err)
			}
			if err := os.Remove(filepath.Join(tmpDir, "coverflex_refresh_token.txt")); err != nil {
				log.Printf("Warning: failed to remove refresh token file: %v", err)
			}
		}
	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("An HTTP error occurred: %d\nResponse: %s", resp.StatusCode, string(bodyBytes))
	}
}

func login() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Enter your password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Step 1: Get OTP challenge
	fmt.Println("Requesting OTP...")
	payload := sessionRequest{Email: email, Password: password}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error creating JSON payload: %v", err)
	}

	req, err := http.NewRequest("POST", sessionURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error during OTP request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusAccepted { // 202
		var otpResp otpResponse
		if err := json.NewDecoder(resp.Body).Decode(&otpResp); err != nil {
			log.Fatalf("Error decoding OTP response: %v", err)
		}
		fmt.Printf("OTP sent to phone ending in ...%s\n", otpResp.PhoneLastDigits)
	} else {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Fatalf("Unexpected status code during OTP request: %d\nResponse: %s", resp.StatusCode, string(bodyBytes))
	}

	// Step 2: Submit OTP and get tokens
	fmt.Print("Enter the OTP you received: ")
	otp, _ := reader.ReadString('\n')
	payload.OTP = strings.TrimSpace(otp)

	fmt.Println("Submitting OTP...")
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error creating OTP payload: %v", err)
	}

	req, err = http.NewRequest("POST", sessionURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Fatalf("Error creating token request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("Error during token request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusCreated { // 201
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Fatalf("Unexpected status code during token request: %d\nResponse: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokens tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		log.Fatalf("Error decoding token response: %v", err)
	}

	authToken := tokens.Token
	refreshToken := tokens.RefreshToken

	if authToken == "" {
		log.Fatal("Failed to retrieve auth token.")
	}

	fmt.Println("Successfully authenticated.")

	// Step 3: Trust the user agent
	fmt.Println("Trusting this device...")
	req, err = http.NewRequest("POST", trustURL, nil)
	if err != nil {
		log.Fatalf("Error creating trust request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err = client.Do(req)
	if err != nil {
		log.Printf("Warning: Error trusting device: %v", err)
	} else {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("Warning: failed to close response body: %v", err)
			}
		}()
		if resp.StatusCode == http.StatusCreated { // 201
			fmt.Println("Device trusted successfully.")
			var newTokens tokenResponse
			if err := json.NewDecoder(resp.Body).Decode(&newTokens); err == nil {
				authToken = newTokens.Token
				refreshToken = newTokens.RefreshToken
				if newTokens.UserAgentToken != "" {
					fmt.Println("Received user agent token for long-term session.")
				}
			}
		}
	}

	// Step 4: Save tokens to files
	tmpDir := os.TempDir()
	tokenPath := filepath.Join(tmpDir, "coverflex_token.txt")
	refreshTokenPath := filepath.Join(tmpDir, "coverflex_refresh_token.txt")

	if err := os.WriteFile(tokenPath, []byte(authToken), 0600); err != nil {
		log.Fatalf("Error saving auth token: %v", err)
	}
	fmt.Printf("Auth token saved to %s\n", tokenPath)

	if refreshToken != "" {
		if err := os.WriteFile(refreshTokenPath, []byte(refreshToken), 0600); err != nil {
			log.Fatalf("Error saving refresh token: %v", err)
		}
		fmt.Printf("Refresh token saved to %s\n", refreshTokenPath)
	}

	// Step 5: Fetch operations after logging in
	if authToken != "" && refreshToken != "" {
		getOperations(authToken, refreshToken)
	}
}

func main() {
	login()
}