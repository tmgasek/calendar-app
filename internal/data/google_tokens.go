package data

import (
	"database/sql"

	"golang.org/x/oauth2"
)

type GoogleTokenModel struct {
	DB *sql.DB
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
