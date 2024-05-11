package main

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/tmgasek/calendar-app/internal/assert"
)

func TestAuthedGroupView(t *testing.T) {
	// Create a new instance of our application struct which uses the mocked
	// dependencies.
	app := newTestApplication(t)
	// Establish a new test server for running end-to-end tests.
	ts := newTestServer(t, app.sessionManager.LoadAndSave(app.mockAuthentication(app.routes())))
	defer ts.Close()
	// Set up some table-driven tests to check the responses sent by our
	// application for different URLs.
	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid ID",
			urlPath:  "/groups/view/1",
			wantCode: http.StatusOK,
			wantBody: "Test Group",
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/groups/view/2",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative ID",
			urlPath:  "/groups/view/-1",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/groups/view/1.23",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "String ID",
			urlPath:  "/groups/view/abc",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Empty ID",
			urlPath:  "/groups/view/",
			wantCode: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			assert.Equal(t, code, tt.wantCode)
			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}

func TestGroupCreate(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.sessionManager.LoadAndSave(app.mockAuthentication(app.routes())))
	defer ts.Close()

	_, _, body := ts.get(t, "/groups")
	validCSRFToken := extractCSRFToken(t, body)

	const (
		validName = "Test Group"
		validDesc = "Test Description"
		formTag   = "<form action='/groups/create' method='POST' novalidate>"
	)

	tests := []struct {
		name        string
		groupName   string
		groupDesc   string
		csrfToken   string
		wantCode    int
		wantFormTag string
	}{
		{
			name:      "Valid submission",
			groupName: validName,
			groupDesc: validDesc,
			csrfToken: validCSRFToken,
			wantCode:  http.StatusSeeOther,
		},
		{
			name:      "Invalid CSRF Token",
			groupName: validName,
			groupDesc: validDesc,
			csrfToken: "wrongToken",
			wantCode:  http.StatusBadRequest,
		},
		{
			name:      "Empty name",
			groupName: "",
			groupDesc: validDesc,
			csrfToken: validCSRFToken,
			wantCode:  http.StatusUnprocessableEntity,
		},
		{
			name:      "Name too long",
			groupName: "This group name is longer than fifty characters and should be rejected by the handlers",
			groupDesc: validDesc,
			csrfToken: validCSRFToken,
			wantCode:  http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("csrf_token", tt.csrfToken)
			form.Add("name", tt.groupName)
			form.Add("description", tt.groupDesc)

			code, _, body := ts.postForm(t, "/groups", form)
			assert.Equal(t, code, tt.wantCode)
			if tt.wantFormTag != "" {
				assert.StringContains(t, body, tt.wantFormTag)
			}
		})
	}
}
