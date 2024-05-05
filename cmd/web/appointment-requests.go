package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
	"github.com/tmgasek/calendar-app/internal/providers"
)

func (app *application) createAppointmentRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("createAppointment")
	// Get the authenticated user ID
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	// Get the target user ID from the URL
	targetUserID, err := app.readIDParam(r)

	var form appointmentCreateForm

	err = app.decodePostForm(r, &form)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	targetUser, err := app.models.Users.Get(int(targetUserID))
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
		TargetUserID: int(targetUserID),
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

	app.sessionManager.Put(r.Context(), "flash", "Appointment request sent!")
	http.Redirect(w, r, fmt.Sprintf("/users/profile/%d", targetUserID), http.StatusSeeOther)
}

func (app *application) viewAppointmentRequests(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	appointmentRequests, err := app.models.AppointmentRequests.GetForUser(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.AppointmentRequests = appointmentRequests
	app.render(w, http.StatusOK, "appointment-requests.tmpl", data)
}

func (app *application) updateAppointmentRequest(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	fmt.Printf("userID: %v\n", userID)

	requestID, err := app.readIDParam(r)
	fmt.Printf("requestID: %v\n", requestID)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	action := r.FormValue("action")
	if action != "confirmed" && action != "declined" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	fmt.Printf("action: %v\n", action)

	if action == "declined" {
		app.infoLog.Println("Declining appointment request")
		err := app.models.AppointmentRequests.Delete(int(requestID))
		if err != nil {
			app.serverError(w, err)
			return
		}
		http.Redirect(w, r, "/requests", http.StatusSeeOther)
		return
	}

	// Retrieve the appointment request from the database
	request, err := app.models.AppointmentRequests.Get(int(requestID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Printf("request: %v\n", request)

	// Create a new event struct.
	newEventData := providers.NewEventData{
		Title:       request.Title,
		Description: request.Description,
		StartTime:   request.StartTime,
		EndTime:     request.EndTime,
		Location:    request.Location,
	}

	newAppointment := &data.Appointment{
		CreatorID:   request.RequesterID,
		TargetID:    request.TargetUserID,
		Title:       request.Title,
		Description: request.Description,
		StartTime:   request.StartTime,
		EndTime:     request.EndTime,
		Location:    request.Location,
	}

	// Save the appointment to the database
	newAppointmentID, err := app.models.Appointments.Insert(newAppointment)
	if err != nil || newAppointmentID == 0 {
		app.serverError(w, err)
		return
	}

	// Create a new slice of userIDs containing the IDs of both the requester and the target.
	userIDs := []int{request.RequesterID, userID}

	// Process appointments for both users.
	for _, userID := range userIDs {
		linkedProviders, err := providers.GetLinkedProviders(userID, &app.models, app.googleOAuthConfig, app.azureOAuth2Config)
		if err != nil {
			app.serverError(w, err)
			return
		}

		for _, p := range linkedProviders {
			app.infoLog.Printf("Creating event from provider %s for user %d\n", p.Name(), userID)

			client, err := providers.GetClient(p, userID, &app.models)
			if err != nil {
				app.serverError(w, err)
				return
			}
			eventID, err := p.CreateEvent(userID, client, newEventData)
			if err != nil {
				app.errorLog.Fatalf("Error creating event: %v\n", err)
				return
			}

			appointmentEvent := &data.AppointmentEvent{
				AppointmentID:   newAppointmentID,
				UserID:          userID,
				ProviderName:    p.Name(),
				ProviderEventID: eventID,
			}
			err = app.models.AppointmentEvents.Insert(appointmentEvent)
			if err != nil {
				app.serverError(w, err)
				return
			}

			app.infoLog.Printf("Provider: %s, Event ID: %s\n", p.Name(), eventID)
		}
	}

	// Delete the appointment request
	err = app.models.AppointmentRequests.Delete(int(requestID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Appointment confirmed!")
	http.Redirect(w, r, "/requests", http.StatusSeeOther)
}
