package main

import (
	"net/http"
	"testing"

	"github.com/tmgasek/calendar-app/internal/assert"
)

func TestAuthedSettingsPage(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.sessionManager.LoadAndSave(app.mockAuthentication(app.routes())))

	defer ts.Close()

	code, _, _ := ts.get(t, "/settings")

	assert.Equal(t, code, http.StatusOK)
}
