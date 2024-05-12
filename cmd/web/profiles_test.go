package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/tmgasek/calendar-app/internal/assert"
	"github.com/tmgasek/calendar-app/internal/data"
)

func TestAuthedUserProfile(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.sessionManager.LoadAndSave(app.mockAuthentication(app.routes())))

	defer ts.Close()

	code, _, _ := ts.get(t, "/users/profile")

	assert.Equal(t, code, http.StatusOK)
}

func TestUnauthedUserProfile(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	defer ts.Close()

	code, _, _ := ts.get(t, "/users/profile")

	assert.Equal(t, code, http.StatusSeeOther)
}

func TestViewOtherUserProfile(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.sessionManager.LoadAndSave(app.mockAuthentication(app.routes())))
	defer ts.Close()

	code, _, _ := ts.get(t, "/users/profile/2")
	assert.Equal(t, code, http.StatusOK)
}

func TestViewInvalidUserProfile(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.sessionManager.LoadAndSave(app.mockAuthentication(app.routes())))
	defer ts.Close()

	code, _, _ := ts.get(t, "/users/profile/invalid")
	assert.Equal(t, code, http.StatusBadRequest)
}

func TestRedirectToOwnProfile(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.sessionManager.LoadAndSave(app.mockAuthentication(app.routes())))
	defer ts.Close()

	code, _, _ := ts.get(t, "/users/profile/1")
	assert.Equal(t, code, http.StatusSeeOther)
}

func TestInitHourlyAvailability(t *testing.T) {
	testCases := []struct {
		name     string
		start    time.Time
		end      time.Time
		events   []*data.Event
		expected []HourlyAvailability
	}{
		{
			name:   "No events",
			start:  time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
			end:    time.Date(2023, 6, 3, 0, 0, 0, 0, time.UTC),
			events: []*data.Event{},
			expected: []HourlyAvailability{
				{Date: "2023-06-01", Hours: [24]string{"free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free"}},
				{Date: "2023-06-02", Hours: [24]string{"free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free"}},
			},
		},
		{
			name:  "Single event",
			start: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2023, 6, 3, 0, 0, 0, 0, time.UTC),
			events: []*data.Event{
				{StartTime: time.Date(2023, 6, 2, 10, 0, 0, 0, time.UTC), EndTime: time.Date(2023, 6, 2, 12, 0, 0, 0, time.UTC)},
			},
			expected: []HourlyAvailability{
				{Date: "2023-06-01", Hours: [24]string{"free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free"}},
				{Date: "2023-06-02", Hours: [24]string{"free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "busy", "busy", "busy", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free", "free"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := &application{}
			result := app.initHourlyAvailability(tc.start, tc.end, tc.events)
			for i := range result {
				assert.Equal(t, tc.expected[i].Date, result[i].Date)
				for j := range result[i].Hours {
					assert.Equal(t, tc.expected[i].Hours[j], result[i].Hours[j])
				}
			}
		})
	}
}
