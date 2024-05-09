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
	Users               UserModelInterface
	AuthTokens          AuthTokenModelInterface
	Appointments        AppointmentModelInterface
	AppointmentRequests AppointmentRequestModelInterface
	Events              EventModelInterface
	AppointmentEvents   AppointmentEventModelInterface
	Groups              GroupModelInterface
}

// For ease of use
func NewModels(db *sql.DB) Models {
	return Models{
		Users:               &UserModel{DB: db},
		AuthTokens:          &AuthTokenModel{DB: db},
		Appointments:        &AppointmentModel{DB: db},
		AppointmentRequests: &AppointmentRequestModel{DB: db},
		Events:              &EventModel{DB: db},
		AppointmentEvents:   &AppointmentEventModel{DB: db},
		Groups:              &GroupModel{DB: db},
	}
}
