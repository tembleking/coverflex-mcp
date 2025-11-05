package domain

// TokenPair holds the access and refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// TokenRepository defines the interface for token persistence.
type TokenRepository interface {
	GetTokens() (*TokenPair, error)
	SaveTokens(accessToken, refreshToken string) error
	DeleteTokens() error
}
