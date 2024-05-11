package data

import (
	"testing"
	"time"

	"github.com/tmgasek/calendar-app/internal/assert"
)

func TestAppointmentRequestModelInsert(t *testing.T) {
	db := newTestDB(t)
	m := AppointmentRequestModel{DB: db}

	request := &AppointmentRequest{
		RequesterID:  1,
		TargetUserID: 2,
		Title:        "Test Request From Test",
		Description:  "This is a test request",
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(time.Hour),
		Location:     "Test Location",
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		TimeZone:     "UTC",
	}

	err := m.Insert(request)
	assert.NilError(t, err)

	// Check if the request record is inserted correctly
	var count int
	// Get count of record that have title "Test Request From Test"
	err = db.QueryRow("SELECT COUNT(*) FROM appointment_requests WHERE title = $1", "Test Request From Test").Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, count, 1)
}

func TestAppointmentRequestModelGetForUser(t *testing.T) {
	db := newTestDB(t)
	m := AppointmentRequestModel{DB: db}

	userID := 2

	requests, err := m.GetForUser(userID)
	assert.NilError(t, err)
	assert.Equal(t, len(requests), 2)

	assert.Equal(t, requests[0].RequestID, 1)
	assert.Equal(t, requests[0].RequesterID, 1)
	assert.Equal(t, requests[0].TargetUserID, 2)
	assert.Equal(t, requests[0].Title, "Request 1")
	assert.Equal(t, requests[0].Requester.Name, "Alice")
	assert.Equal(t, requests[0].Requester.Email, "alice@example.com")

	assert.Equal(t, requests[1].RequestID, 2)
	assert.Equal(t, requests[1].RequesterID, 1)
	assert.Equal(t, requests[1].TargetUserID, 2)
	assert.Equal(t, requests[1].Title, "Request 2")
	assert.Equal(t, requests[1].Requester.Name, "Alice")
	assert.Equal(t, requests[1].Requester.Email, "alice@example.com")
}

func TestAppointmentRequestModelGet(t *testing.T) {
	db := newTestDB(t)
	m := AppointmentRequestModel{DB: db}

	requestID := 1

	request, err := m.Get(requestID)
	assert.NilError(t, err)
	assert.NotNil(t, request)

	assert.Equal(t, request.RequestID, 1)
	assert.Equal(t, request.RequesterID, 1)
	assert.Equal(t, request.TargetUserID, 2)
	assert.Equal(t, request.Title, "Request 1")
	assert.Equal(t, request.Requester.Name, "Alice")
	assert.Equal(t, request.Requester.Email, "alice@example.com")
}

func TestAppointmentRequestModelDelete(t *testing.T) {
	db := newTestDB(t)
	m := AppointmentRequestModel{DB: db}

	requestID := 1

	err := m.Delete(requestID)
	assert.NilError(t, err)

	// Check if the request record is deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM appointment_requests WHERE request_id = $1", requestID).Scan(&count)
	assert.NilError(t, err)
	assert.Equal(t, count, 0)
}
