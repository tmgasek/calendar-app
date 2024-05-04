package main

import (
	"fmt"
	"net/http"

	"github.com/tmgasek/calendar-app/internal/data"
	"github.com/tmgasek/calendar-app/internal/providers"
)

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

	// Create appointments for both users.
	appointments := []*data.Appointment{
		{
			UserID:      userID,
			Title:       request.Title,
			Description: request.Description,
			StartTime:   request.StartTime,
			EndTime:     request.EndTime,
			Location:    request.Location,
		},
		{
			UserID:      request.RequesterID,
			Title:       request.Title,
			Description: request.Description,
			StartTime:   request.StartTime,
			EndTime:     request.EndTime,
			Location:    request.Location,
		},
	}

	// Process appointments for both users.
	for _, appointment := range appointments {
		providers, err := providers.GetLinkedProviders(appointment.UserID, &app.models, app.googleOAuthConfig, app.azureOAuth2Config)
		if err != nil {
			app.serverError(w, err)
			return
		}

		for _, provider := range providers {
			app.infoLog.Printf("Creating event from provider %s for user %d\n", provider.Name(), appointment.UserID)

			token, err := app.models.AuthTokens.Token(appointment.UserID, provider.Name())
			if err != nil {
				app.serverError(w, err)
				return
			}

			client := provider.CreateClient(r.Context(), token)
			eventID, err := provider.CreateEvent(appointment.UserID, client, newEventData)
			if err != nil {
				app.errorLog.Fatalf("Error creating event: %v\n", err)
				return
			}

			if provider.Name() == "google" {
				appointment.GoogleEventID = eventID
			} else if provider.Name() == "microsoft" {
				appointment.MicrosoftEventID = eventID
			}

			app.infoLog.Printf("Provider: %s, Event ID: %s\n", provider.Name(), eventID)
		}

		// Save the appointment to the database
		err = app.models.Appointments.Insert(appointment)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	// Delete the appointment request
	err = app.models.AppointmentRequests.Delete(int(requestID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/requests", http.StatusSeeOther)
}
