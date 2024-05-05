package main

import (
	"fmt"
	"net/http"

	"github.com/tmgasek/calendar-app/internal/providers"
	"github.com/tmgasek/calendar-app/internal/validator"
)

// include struct tags to tell the decoder how to map HTML form vals to
// struct fields. "-" tells it to ignore a field!
type appointmentCreateForm struct {
	Title               string `form:"title"`
	Description         string `form:"description"`
	StartTime           string `form:"start_time"`
	EndTime             string `form:"end_time"`
	Location            string `form:"location"`
	validator.Validator `form:"-"`
}

func (app *application) deleteAppointment(w http.ResponseWriter, r *http.Request) {
	appointmentID, err := app.readIDParam(r)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	currUserID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	fmt.Println("appointmentID: ", appointmentID)
	fmt.Println("userID: ", currUserID)

	// Get the appointment from the database.
	appointment, err := app.models.Appointments.Get(int(appointmentID))

	// Check if userID is the creator or target of the appointment
	if appointment.CreatorID != currUserID && appointment.TargetID != currUserID {
		app.clientError(w, http.StatusForbidden)
		return
	}

	// Delete from both users calendars.
	appointmentEvents, err := app.models.AppointmentEvents.GetByAppointmentID(int(appointmentID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	for _, event := range appointmentEvents {
		provider, err := providers.GetProviderByName(event.UserID, event.ProviderName, &app.models, app.googleOAuthConfig, app.azureOAuth2Config)
		if err != nil {
			app.serverError(w, err)
			return
		}

		client, err := providers.GetClient(provider, event.UserID, &app.models)
		if err != nil {
			app.serverError(w, err)
			return
		}

		err = provider.DeleteEvent(event.UserID, client, event.ProviderName, event.ProviderEventID)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	// Delete the appointment from the database.
	err = app.models.Appointments.Delete(int(appointmentID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Appointment successfully deleted!")
	// Redirect back to the profile page
	http.Redirect(w, r, "/appointments", http.StatusSeeOther)
}

func (app *application) viewAppointments(w http.ResponseWriter, r *http.Request) {
	templateData := app.newTemplateData(r)
	// Get the authenticated user ID
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	// Get the user's appointments
	appointments, err := app.models.Appointments.GetForUser(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Render the appointments
	templateData.Appointments = appointments
	app.render(w, http.StatusOK, "appointments.tmpl", templateData)
}
