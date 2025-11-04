package fs

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/tembleking/coverflex-mcp/domain"
)

const (
	tokenFileName        = "coverflex_token.txt"
	refreshTokenFileName = "coverflex_refresh_token.txt"
)

// TokenRepository handles token persistence in the filesystem.
type TokenRepository struct {
	tokenPath        string
	refreshTokenPath string
}

// NewTokenRepository creates a new filesystem token repository.
func NewTokenRepository() *TokenRepository {
	tmpDir := os.TempDir()
	return &TokenRepository{
		tokenPath:        filepath.Join(tmpDir, tokenFileName),
		refreshTokenPath: filepath.Join(tmpDir, refreshTokenFileName),
	}
}

// GetTokens retrieves the tokens from the filesystem.
func (r *TokenRepository) GetTokens() (*domain.TokenPair, error) {
	_, errToken := os.Stat(r.tokenPath)
	_, errRefreshToken := os.Stat(r.refreshTokenPath)

	if os.IsNotExist(errToken) || os.IsNotExist(errRefreshToken) {
		return nil, fmt.Errorf("token files not found")
	}

	tokenBytes, err := os.ReadFile(r.tokenPath)
	if err != nil {
		return nil, fmt.Errorf("could not read token file: %w", err)
	}

	refreshBytes, err := os.ReadFile(r.refreshTokenPath)
	if err != nil {
		return nil, fmt.Errorf("could not read refresh token file: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  strings.TrimSpace(string(tokenBytes)),
		RefreshToken: strings.TrimSpace(string(refreshBytes)),
	}, nil
}

// SaveTokens saves the tokens to the filesystem.
func (r *TokenRepository) SaveTokens(accessToken, refreshToken string) error {
	if err := os.WriteFile(r.tokenPath, []byte(accessToken), 0600); err != nil {
		return fmt.Errorf("error saving auth token: %w", err)
	}
	slog.Info("Auth token saved", "path", r.tokenPath)

	if refreshToken != "" {
		if err := os.WriteFile(r.refreshTokenPath, []byte(refreshToken), 0600); err != nil {
			return fmt.Errorf("error saving refresh token: %w", err)
		}
		slog.Info("Refresh token saved", "path", r.refreshTokenPath)
	}
	return nil
}

// DeleteTokens removes the token files from the filesystem.
func (r *TokenRepository) DeleteTokens() error {
	if err := os.Remove(r.tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token file: %w", err)
	}
	if err := os.Remove(r.refreshTokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove refresh token file: %w", err)
	}
	return nil
}
