package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

type AuthTokenModel struct {
	DB *sql.DB
}

type AuthToken struct {
	UserID       int
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
	Scope        string
	AuthProvider string
}

func (m *AuthTokenModel) SaveToken(userID int, authProvider string, token *oauth2.Token) error {
	query := `
        INSERT INTO auth_tokens (
			user_id,
            auth_provider,
            access_token,
            refresh_token,
            token_type,
            expiry,
            scope
        ) 
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (user_id, auth_provider) 
        DO UPDATE SET access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token, token_type = EXCLUDED.token_type, expiry = EXCLUDED.expiry, scope = EXCLUDED.scope;
    `

	_, err := m.DB.Exec(
		query,
		userID,
		authProvider,
		token.AccessToken,
		token.RefreshToken,
		token.TokenType,
		token.Expiry,
		token.Extra("scope"),
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *AuthTokenModel) Token(userID int, authProvider string) (*oauth2.Token, error) {
	var token AuthToken
	query := `SELECT access_token, refresh_token, token_type, expiry FROM auth_tokens WHERE user_id = $1 AND auth_provider = $2`
	row := m.DB.QueryRow(query, userID, authProvider)
	err := row.Scan(&token.AccessToken, &token.RefreshToken, &token.TokenType, &token.Expiry)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("No %s token found for user %d\n", authProvider, userID)
			return nil, nil
		}
		return nil, err
	}
	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}, nil
}
