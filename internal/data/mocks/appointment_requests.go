package mocks

import (
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
)

type AppointmentRequestModel struct{}

var mockAppointmentRequest = &data.AppointmentRequest{
	RequestID:       1,
	Title:           "Test Appointment",
	Description:     "Test Description",
	StartTime:       time.Date(2021, 1, 1, 11, 0, 0, 0, time.UTC),
	EndTime:         time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC),
	Location:        "Test Location",
	Status:          "confirmed",
	CreatedAt:       time.Now(),
	UpdatedAt:       time.Now(),
	TimeZone:        "UTC",
	RequesterID:     1,
	AppointmentType: "individual",
	GroupID:         0,
	TargetUserID:    2,
}

func (m *AppointmentRequestModel) Insert(request *data.AppointmentRequest) error {
	return nil
}

func (m *AppointmentRequestModel) GetForUser(userID int) ([]*data.AppointmentRequest, error) {
	return []*data.AppointmentRequest{mockAppointmentRequest}, nil
}

func (m *AppointmentRequestModel) Get(requestID int) (*data.AppointmentRequest, error) {
	return mockAppointmentRequest, nil
}

func (m *AppointmentRequestModel) Delete(requestID int) error {
	return nil
}
