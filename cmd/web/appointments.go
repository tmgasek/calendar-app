package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
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
	TargetUserID        int64  `form:"target_user_id"`
	validator.Validator `form:"-"`
}

func (app *application) createAppointment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("createAppointment")
	// Get the authenticated user ID
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	var form appointmentCreateForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	targetUser, err := app.models.Users.Get(int(form.TargetUserID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Parse the start and end times
	startTime, err := time.Parse("2006-01-02T15:04", form.StartTime)
	if err != nil {
		app.serverError(w, err)
		return
	}
	endTime, err := time.Parse("2006-01-02T15:04", form.EndTime)
	if err != nil {
		app.serverError(w, err)
		return
	}

	requestee, err := app.models.Users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	type EmailData struct {
		RequesteeName string
	}

	// Create the appointment request.
	appointmentRequest := &data.AppointmentRequest{
		RequesterID:  int(userID),
		TargetUserID: int(form.TargetUserID),
		Title:        form.Title,
		Description:  form.Description,
		StartTime:    startTime,
		EndTime:      endTime,
		Location:     form.Location,
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = app.models.AppointmentRequests.Insert(appointmentRequest)
	if err != nil {
		app.errorLog.Println(err)
		app.serverError(w, err)
		return
	}

	emailData := EmailData{
		RequesteeName: requestee.Name,
	}

	err = app.mailer.Send(targetUser.Email, "confirm-appointment.tmpl", emailData)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.infoLog.Println("********** Email sent")
}

func (app *application) deleteAppointment(w http.ResponseWriter, r *http.Request) {
	// Get the event ID from the URL parameters.
	appointmentID := r.FormValue("appointment_id")
	googleEventID := r.FormValue("google_event_id")
	microsoftEventID := r.FormValue("microsoft_event_id")

	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	fmt.Println("appointmentID: ", appointmentID)
	fmt.Println("googleEventID: ", googleEventID)
	fmt.Println("microsoftEventID: ", microsoftEventID)
	fmt.Println("userID: ", userID)

	if appointmentID == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	providers, err := providers.GetLinkedProviders(userID, &app.models, app.googleOAuthConfig, app.azureOAuth2Config)
	if err != nil {
		app.serverError(w, err)
		return
	}

	for _, provider := range providers {
		token, err := app.models.AuthTokens.Token(userID, provider.Name())
		if err != nil {
			app.serverError(w, err)
			return
		}

		client := provider.CreateClient(r.Context(), token)

		if provider.Name() == "google" {
			err = provider.DeleteEvent(userID, client, provider.Name(), googleEventID)
		} else if provider.Name() == "microsoft" {
			err = provider.DeleteEvent(userID, client, provider.Name(), microsoftEventID)
		}

		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	// Convert appointmentID to int.
	appointmentIDInt, err := strconv.Atoi(appointmentID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	// Delete the appointment from the database.
	err = app.models.Appointments.Delete(appointmentIDInt)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Redirect back to the profile page
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}

func (app *application) viewAppointments(w http.ResponseWriter, r *http.Request) {
	templateData := app.newTemplateData(r)
	// Get the authenticated user ID
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	// Get the user's appointments
	appointments, err := app.models.Appointments.GetByUserID(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Render the appointments
	templateData.Appointments = appointments
	app.render(w, http.StatusOK, "appointments.tmpl", templateData)
}
