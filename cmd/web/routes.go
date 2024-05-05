package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/tmgasek/calendar-app/ui"
)

/*
   Home and Static Content:
       GET /: Home page.
       GET /static/*filepath: Serve static files.

   User Authentication:
       GET /signup: Display signup form.
       POST /signup: Process signup form.
       GET /login: Display login form.
       POST /login: Process login form.
       POST /logout: Handle logout action.

   User Profile Management:
       GET /profile: Display user's profile.
       GET /profile/edit: Display form to edit profile.
       POST /profile/edit: Process profile edit form.

   Calendar Integration (OAuth):
       GET /oauth/google/link: Display page to link Google account.
       GET /oauth/google/callback: Handle Google OAuth callback.
       GET /oauth/microsoft/link: Display page to link Microsoft account.
       GET /oauth/microsoft/callback: Handle Microsoft OAuth callback.

   Event Management:
       GET /events: Display list of user's events.
       GET /events/new: Display form to create a new event.
       POST /events/new: Process new event form.
       GET /events/google: Display Google events.
       GET /events/microsoft: Display Microsoft/Outlook events.

   Viewing Other Users' Calendars:
       GET /users/:id/calendar: View another user's calendar.

   Appointment Booking:
       GET /appointments/new: Display form to book an appointment.
       POST /appointments/new: Process appointment booking form.
       GET /appointments/:id: View appointment details.
*/

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
	router.Handler(http.MethodGet, "/users/:id/calendar", protected.ThenFunc(app.viewUserProfile))

	// Appointments
	router.Handler(http.MethodGet, "/appointments", protected.ThenFunc(app.viewAppointments))
	router.Handler(http.MethodPost, "/appointments/create", protected.ThenFunc(app.createAppointment))
	router.Handler(http.MethodPost, "/appointments/delete", protected.ThenFunc(app.deleteAppointment))

	// Appointment Requests
	router.Handler(http.MethodGet, "/requests", protected.ThenFunc(app.viewAppointmentRequests))

	router.Handler(http.MethodPost, "/request/:id/update", protected.ThenFunc(app.updateAppointmentRequest))

	// Create a new middleware chain.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
