package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

type GoogleTokenModel struct {
	DB *sql.DB
}

type GoogleToken struct {
	UserID       int
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
	Scope        string
}

func (m *GoogleTokenModel) SaveToken(userID int, token *oauth2.Token) error {
	query := `
        INSERT INTO google_tokens (user_id, access_token, refresh_token, token_type, expiry, scope) 
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (user_id) 
        DO UPDATE SET access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token, token_type = EXCLUDED.token_type, expiry = EXCLUDED.expiry, scope = EXCLUDED.scope;
    `

	_, err := m.DB.Exec(query, userID, token.AccessToken, token.RefreshToken, token.TokenType, token.Expiry, token.Extra("scope"))
	if err != nil {
		return err
	}
	return nil
}

func (m *GoogleTokenModel) Token(userID int) (*oauth2.Token, error) {
	var token GoogleToken
	query := `SELECT access_token, refresh_token, token_type, expiry FROM google_tokens WHERE user_id = $1`
	row := m.DB.QueryRow(query, userID)
	err := row.Scan(&token.AccessToken, &token.RefreshToken, &token.TokenType, &token.Expiry)
	if err != nil {
		return nil, err
	}
	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}, nil
}

func (m *GoogleTokenModel) Expired(token *oauth2.Token) bool {
	return token.Expiry.Before(time.Now())
}

func (m *GoogleTokenModel) RefreshGoogleToken(userID int, config *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	fmt.Printf("token: %v\n", token)
	if !token.Valid() {
		newToken, err := config.TokenSource(context.Background(), token).Token()

		fmt.Printf("newToken: %v\n", newToken)

		if err != nil {
			// Does this mean the expiry token is invalid? Unable to refresh token.
			// TODO: redirect to auth again?
			return nil, err
		}

		// Save the new token to the database.
		err = m.SaveToken(userID, newToken)
		if err != nil {
			return nil, err
		}

		return newToken, nil
	}

	return token, nil
}
