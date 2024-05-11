package data

import (
	"testing"

	"github.com/tmgasek/calendar-app/internal/assert"
)

func TestAppointmentEventModelInsert(t *testing.T) {
	db := newTestDB(t)
	m := AppointmentEventModel{DB: db}

	event := &AppointmentEvent{
		AppointmentID:   1,
		UserID:          1,
		ProviderName:    "google",
		ProviderEventID: "event_1_from_test",
	}

	err := m.Insert(event)
	assert.NilError(t, err)

	// Check if the event record is inserted correctly
	var count int
	// Check the ProviderEventID is "event_1_from_test"
	err = db.QueryRow("SELECT COUNT(*) FROM appointment_events WHERE provider_event_id = $1", "event_1_from_test").Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, count, 1)
}

func TestAppointmentEventModelGetByAppointmentID(t *testing.T) {
	db := newTestDB(t)
	m := AppointmentEventModel{DB: db}

	appointmentID := 1

	events, err := m.GetByAppointmentID(appointmentID)
	assert.NilError(t, err)
	assert.Equal(t, len(events), 2)

	assert.Equal(t, events[0].ID, 1)
	assert.Equal(t, events[0].AppointmentID, 1)
	assert.Equal(t, events[0].UserID, 1)
	assert.Equal(t, events[0].ProviderName, "google")
	assert.Equal(t, events[0].ProviderEventID, "event_1")

	assert.Equal(t, events[1].ID, 2)
	assert.Equal(t, events[1].AppointmentID, 1)
	assert.Equal(t, events[1].UserID, 2)
	assert.Equal(t, events[1].ProviderName, "outlook")
	assert.Equal(t, events[1].ProviderEventID, "event_2")
}
