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
	Users               UserModel
	AuthTokens          AuthTokenModel
	Events              EventModel
	AppointmentRequests AppointmentRequestModel
}

// For ease of use
func NewModels(db *sql.DB) Models {
	return Models{
		Users:               UserModel{DB: db},
		AuthTokens:          AuthTokenModel{DB: db},
		Events:              EventModel{DB: db},
		AppointmentRequests: AppointmentRequestModel{DB: db},
	}
}
