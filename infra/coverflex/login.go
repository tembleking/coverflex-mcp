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

	resp, err := c.httpClient.Do(req)
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

	resp, err = c.httpClient.Do(req)
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

	resp, err = c.httpClient.Do(req)
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
	if err := c.tokenRepo.SaveTokens(authToken, refreshToken); err != nil {
		log.Fatalf("Error saving tokens: %v", err)
	}

	// Step 5: Fetch operations after logging in
	if authToken != "" && refreshToken != "" {
		c.GetOperations(authToken, refreshToken)
	}
}
