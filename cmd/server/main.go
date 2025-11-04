package main

func main() {
	Execute()
}

// func main() {
// 	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
// 	slog.SetDefault(logger)

// 	tokenRepo := fs.NewTokenRepository()
// 	client := coverflex.NewClient(tokenRepo)

// 	// --- Flag Definitions ---
// 	user := flag.String("user", "", "Your Coverflex email")
// 	pass := flag.String("pass", "", "Your Coverflex password")
// 	otp := flag.String("otp", "", "The OTP you received via SMS")
// 	forceRefresh := flag.Bool("force-refresh", false, "Force a refresh of the authentication tokens")

// 	flag.Parse()

// 	// --- Logic based on flags ---

// 	// Case 1: Force Refresh
// 	if *forceRefresh {
// 		slog.Info("Force refresh option detected.")
// 		tokens, err := tokenRepo.GetTokens()
// 		if err != nil {
// 			slog.Error("Refresh token file not found. Cannot force refresh. Please log in first.")
// 			os.Exit(1)
// 		}
// 		newAuthToken, newRefreshToken := client.RefreshTokens(tokens.RefreshToken)
// 		if newAuthToken != "" {
// 			slog.Info("\nTokens have been refreshed. Let's test the new token:")
// 			client.GetOperations(newAuthToken, newRefreshToken)
// 		} else {
// 			slog.Error("Failed to refresh tokens.")
// 			os.Exit(1)
// 		}
// 		return
// 	}

// 	// Case 2: Login with OTP
// 	if *user != "" && *pass != "" && *otp != "" {
// 		slog.Info("User, password, and OTP provided. Attempting to log in...")
// 		if err := client.Login(*user, *pass, *otp); err != nil {
// 			slog.Error("Login failed", "error", err)
// 			os.Exit(1)
// 		}
// 		slog.Info("Logged in.")
// 		return
// 	}

// 	// Case 3: Request OTP
// 	if *user != "" && *pass != "" {
// 		slog.Info("User and password provided. Requesting OTP...")
// 		if err := client.RequestOTP(*user, *pass); err != nil {
// 			slog.Error("Failed to request OTP", "error", err)
// 			os.Exit(1)
// 		}
// 		slog.Info("An OTP has been sent to your phone. Please re-run the command with the --otp flag.")
// 		return
// 	}
// }
