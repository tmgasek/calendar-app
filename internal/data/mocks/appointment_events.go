package mocks

import "github.com/tmgasek/calendar-app/internal/data"

type AppointmentEventModel struct{}

var mockAppointmentEvent = &data.AppointmentEvent{
	ID:            1,
	AppointmentID: 1,
	UserID:        1,
}

func (m *AppointmentEventModel) Insert(a *data.AppointmentEvent) error {
	return nil
}

func (m *AppointmentEventModel) GetByAppointmentID(appointmentID int) ([]*data.AppointmentEvent, error) {
	return []*data.AppointmentEvent{mockAppointmentEvent}, nil
}
