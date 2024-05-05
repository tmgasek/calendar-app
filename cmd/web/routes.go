package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/tmgasek/calendar-app/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// create wrapper around our NotFound() helper.
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	router.HandlerFunc(http.MethodGet, "/ping", ping)

	// For dynamic routes.
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	// Protected application routes.
	protected := dynamic.Append(app.requireAuthentication)
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))

	// Google OAuth routes.
	router.Handler(http.MethodGet, "/oauth/google/link", protected.ThenFunc(app.linkGoogleAccount))
	router.Handler(http.MethodGet, "/oauth/google/callback", protected.ThenFunc(app.handleGoogleCalendarCallback))

	// Microsoft OAuth routes.
	router.Handler(http.MethodGet, "/oauth/microsoft/link", protected.ThenFunc(app.redirectToMicrosoftLogin))
	router.Handler(http.MethodGet, "/oauth/microsoft/callback", protected.ThenFunc(app.handleMicrosoftAuthCallback))

	// Profile views
	router.Handler(http.MethodGet, "/user/profile", protected.ThenFunc(app.userProfile))
	router.Handler(http.MethodGet, "/users/profile/:id", protected.ThenFunc(app.viewUserProfile))
	// User search
	router.Handler(http.MethodGet, "/users/search", protected.ThenFunc(app.searchUsers))

	// Appointments
	router.Handler(http.MethodGet, "/appointments", protected.ThenFunc(app.viewAppointments))
	router.Handler(http.MethodPost, "/appointments/create/:id", protected.ThenFunc(app.createAppointmentRequest))
	router.Handler(http.MethodPost, "/appointments/delete/:id", protected.ThenFunc(app.deleteAppointment))

	// Appointment Requests
	router.Handler(http.MethodGet, "/requests", protected.ThenFunc(app.viewAppointmentRequests))
	router.Handler(http.MethodPost, "/request/:id/update", protected.ThenFunc(app.updateAppointmentRequest))

	// Settings
	router.Handler(http.MethodGet, "/settings", protected.ThenFunc(app.viewSettings))

	// Create a new middleware chain.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
