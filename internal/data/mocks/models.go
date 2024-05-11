package mocks

import (
	"errors"
	"github.com/tmgasek/calendar-app/internal/data"
	"github.com/tmgasek/calendar-app/internal/mailer"
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

type mockMailer struct {
	SendFunc func(recipient, templateFile string, data any) error
}

func (m *mockMailer) Send(recipient, templateFile string, data any) error {
	if m.SendFunc != nil {
		return m.SendFunc(recipient, templateFile, data)
	}
	return nil
}

func NewMockMailer() mailer.MailerInterface {
	return &mockMailer{}
}
