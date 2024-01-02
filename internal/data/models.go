package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound     = errors.New("record not found")
	ErrEditConflict       = errors.New("edit conflict")
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Models struct {
	Users        UserModel
	GoogleTokens GoogleTokenModel
}

// For ease of use
func NewModels(db *sql.DB) Models {
	return Models{
		Users:        UserModel{DB: db},
		GoogleTokens: GoogleTokenModel{DB: db},
	}
}
