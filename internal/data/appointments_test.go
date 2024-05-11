package data

import (
	"testing"
	"time"

	"github.com/tmgasek/calendar-app/internal/assert"
)

func TestAppointmentModelInsert(t *testing.T) {
	db := newTestDB(t)
	m := AppointmentModel{DB: db}

	appointment := &Appointment{
		CreatorID:   1,
		TargetID:    2,
		Title:       "Test Appointment",
		Description: "This is a test appointment",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
		Location:    "Test Location",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		TimeZone:    "UTC",
		Visibility:  "public",
		Recurrence:  "daily",
	}

	id, err := m.Insert(appointment)
	assert.NilError(t, err)
	assert.Greater(t, id, 0)

	// Check if the appointment record is inserted correctly
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM appointments WHERE id = $1", id).Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, count, 1)
}

func TestAppointmentModelGetForUser(t *testing.T) {
	db := newTestDB(t)
	m := AppointmentModel{DB: db}

	userID := 1

	appointments, err := m.GetForUser(userID)
	assert.NilError(t, err)
	assert.Equal(t, len(appointments), 2)

	assert.Equal(t, appointments[0].ID, 1)
	assert.Equal(t, appointments[0].CreatorID, 1)
	assert.Equal(t, appointments[0].TargetID, 2)
	assert.Equal(t, appointments[0].Title, "Appointment 1")

	assert.Equal(t, appointments[1].ID, 2)
	assert.Equal(t, appointments[1].CreatorID, 2)
	assert.Equal(t, appointments[1].TargetID, 1)
	assert.Equal(t, appointments[1].Title, "Appointment 2")
}

func TestAppointmentModelDelete(t *testing.T) {
	db := newTestDB(t)
	m := AppointmentModel{DB: db}

	appointmentID := 1

	err := m.Delete(appointmentID)
	assert.NilError(t, err)

	// Check if the appointment record is deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM appointments WHERE id = $1", appointmentID).Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, count, 0)
}

func TestAppointmentModelGet(t *testing.T) {
	db := newTestDB(t)
	m := AppointmentModel{DB: db}

	appointmentID := 2

	appointment, err := m.Get(appointmentID)
	assert.NilError(t, err)
	assert.NotNil(t, appointment)

	assert.Equal(t, appointment.ID, 2)
	assert.Equal(t, appointment.CreatorID, 2)
	assert.Equal(t, appointment.TargetID, 1)
	assert.Equal(t, appointment.Title, "Appointment 2")
}
