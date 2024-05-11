package data

import (
	// "database/sql"
	"testing"
	"time"

	"github.com/tmgasek/calendar-app/internal/assert"
	"golang.org/x/oauth2"
)

func TestAuthTokenModelSaveToken(t *testing.T) {
	db := newTestDB(t)
	m := AuthTokenModel{DB: db}

	userID := 1
	authProvider := "google"
	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	err := m.SaveToken(userID, authProvider, token)
	assert.NilError(t, err)

	// Check if the auth_token record is inserted or updated correctly
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM auth_tokens WHERE user_id = $1 AND auth_provider = $2", userID, authProvider).Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, count, 1)

	var savedToken AuthToken
	err = db.QueryRow("SELECT access_token, refresh_token, token_type, expiry FROM auth_tokens WHERE user_id = $1 AND auth_provider = $2", userID, authProvider).
		Scan(&savedToken.AccessToken, &savedToken.RefreshToken, &savedToken.TokenType, &savedToken.Expiry)
	assert.NilError(t, err)
	assert.Equal(t, savedToken.AccessToken, token.AccessToken)
	assert.Equal(t, savedToken.RefreshToken, token.RefreshToken)
	assert.Equal(t, savedToken.TokenType, token.TokenType)
}

func TestAuthTokenModelToken(t *testing.T) {
	db := newTestDB(t)
	m := AuthTokenModel{DB: db}

	userID := 1
	authProvider := "google"

	token, err := m.Token(userID, authProvider)
	assert.NilError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, token.AccessToken, "access-token-1")
	assert.Equal(t, token.RefreshToken, "refresh-token-1")
	assert.Equal(t, token.TokenType, "Bearer")

	// Test non-existent token
	userID = 2
	token, err = m.Token(userID, authProvider)
	assert.NilError(t, err)
}
