package main

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/tmgasek/calendar-app/internal/assert"
)

func TestCreateAppointmentRequest(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	app := newTestApplication(t)
	ts := newTestServer(t, app.sessionManager.LoadAndSave(app.mockAuthentication(app.routes())))
	defer ts.Close()

	_, _, body := ts.get(t, "/users/profile/2")
	validCSRFToken := extractCSRFToken(t, body)

	const (
		validTitle       = "Appointment Title"
		validDescription = "Appointment Description"
		validStartTime   = "2023-06-01T10:00"
		validEndTime     = "2023-06-01T11:00"
		validLocation    = "Appointment Location"
		formTag          = "<form action='/appointments/create/2' method='POST' novalidate>"
	)

	tests := []struct {
		name        string
		title       string
		description string
		startTime   string
		endTime     string
		location    string
		groupID     int
		csrfToken   string
		wantCode    int
		wantFormTag string
	}{
		{
			name:        "Valid submission",
			title:       validTitle,
			description: validDescription,
			startTime:   validStartTime,
			endTime:     validEndTime,
			location:    validLocation,
			groupID:     0,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusSeeOther,
		},
		{
			name:        "Invalid CSRF Token",
			title:       validTitle,
			description: validDescription,
			startTime:   validStartTime,
			endTime:     validEndTime,
			location:    validLocation,
			groupID:     0,
			csrfToken:   "wrongToken",
			wantCode:    http.StatusBadRequest,
		},
		{
			name:        "Empty title",
			title:       "",
			description: validDescription,
			startTime:   validStartTime,
			endTime:     validEndTime,
			location:    validLocation,
			groupID:     0,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusUnprocessableEntity,
		},
		{
			name:        "Empty description",
			title:       validTitle,
			description: "",
			startTime:   validStartTime,
			endTime:     validEndTime,
			location:    validLocation,
			groupID:     0,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusUnprocessableEntity,
		},
		{
			name:        "Empty start time",
			title:       validTitle,
			description: validDescription,
			startTime:   "",
			endTime:     validEndTime,
			location:    validLocation,
			groupID:     0,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusUnprocessableEntity,
		},
		{
			name:        "Empty end time",
			title:       validTitle,
			description: validDescription,
			startTime:   validStartTime,
			endTime:     "",
			location:    validLocation,
			groupID:     0,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("title", tt.title)
			form.Add("description", tt.description)
			form.Add("start_time", tt.startTime)
			form.Add("end_time", tt.endTime)
			form.Add("location", tt.location)
			form.Add("group_id", strconv.Itoa(tt.groupID))
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/appointments/create/2", form)
			assert.Equal(t, code, tt.wantCode)

			if tt.wantFormTag != "" {
				assert.StringContains(t, body, tt.wantFormTag)
			}
		})
	}
}
