package mocks

import (
	"errors"
	"github.com/tmgasek/calendar-app/internal/data"
)

var (
	ErrRecordNotFound     = errors.New("record not found")
	ErrEditConflict       = errors.New("edit conflict")
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func NewMockModels() data.Models {
	return data.Models{
		Users:               &UserModel{},
		AuthTokens:          &AuthTokenModel{},
		Appointments:        &AppointmentModel{},
		AppointmentRequests: &AppointmentRequestModel{},
		Events:              &EventModel{},
		AppointmentEvents:   &AppointmentEventModel{},
		Groups:              &GroupModel{},
	}
}
