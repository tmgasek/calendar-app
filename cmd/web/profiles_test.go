package main

import (
	"net/http"
	"testing"

	"github.com/tmgasek/calendar-app/internal/assert"
)

func TestAuthedUserProfile(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.sessionManager.LoadAndSave(app.mockAuthentication(app.routes())))

	defer ts.Close()

	code, _, _ := ts.get(t, "/user/profile")

	assert.Equal(t, code, http.StatusOK)
}

func TestUnauthedUserProfile(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	defer ts.Close()

	code, _, _ := ts.get(t, "/user/profile")

	assert.Equal(t, code, http.StatusSeeOther)
}
