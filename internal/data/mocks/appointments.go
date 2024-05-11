package mocks

import (
	"github.com/tmgasek/calendar-app/internal/data"
	"time"
)

var mockAppointment = &data.Appointment{
	ID:              1,
	CreatorID:       1,
	TargetID:        2,
	Title:           "Test Appointment",
	Description:     "Test Description",
	StartTime:       time.Date(2021, 1, 1, 11, 0, 0, 0, time.UTC),
	EndTime:         time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC),
	Location:        "Test Location",
	Status:          "confirmed",
	CreatedAt:       time.Now(),
	UpdatedAt:       time.Now(),
	TimeZone:        "UTC",
	Visibility:      "public",
	Recurrence:      "none",
	AppointmentType: "test",
}

type AppointmentModel struct{}

func (m *AppointmentModel) Insert(a *data.Appointment) (int, error) {
	return 1, nil
}

func (m *AppointmentModel) GetForUser(userID int) ([]*data.Appointment, error) {
	return []*data.Appointment{mockAppointment}, nil
}

func (m *AppointmentModel) Get(id int) (*data.Appointment, error) {
	switch id {
	case 1:
		return mockAppointment, nil
	default:
		return nil, data.ErrRecordNotFound
	}
}

func (m *AppointmentModel) Update(a *data.Appointment) error {
	return nil
}

func (m *AppointmentModel) Delete(id int) error {
	return nil
}
