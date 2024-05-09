package mocks

import (
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
	"golang.org/x/oauth2"
)

var mockAuthToken = &data.AuthToken{
	UserID:       1,
	AccessToken:  "access-token",
	RefreshToken: "refresh-token",
	TokenType:    "Bearer",
	Expiry:       time.Now().Add(time.Hour),
	Scope:        "read write",
	AuthProvider: "google",
}

type AuthTokenModel struct{}

func (m *AuthTokenModel) SaveToken(userID int, authProvider string, token *oauth2.Token) error {
	return nil
}

func (m *AuthTokenModel) Token(userID int, authProvider string) (*oauth2.Token, error) {
	return nil, nil
}
