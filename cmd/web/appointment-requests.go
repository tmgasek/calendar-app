package main

import (
	"fmt"
	"net/http"

	// "github.com/tmgasek/calendar-app/internal/data"
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

	// Create a new event struct
	newEventData := providers.NewEventData{
		Title:       request.Title,
		Description: request.Description,
		StartTime:   request.StartTime,
		EndTime:     request.EndTime,
		Location:    request.Location,
	}

	// Handle sending requests to user 1's providers (the user accepting the request)
	user1Providers, err := providers.GetLinkedProviders(userID, &app.models, app.googleOAuthConfig, app.azureOAuth2Config)
	if err != nil {
		app.serverError(w, err)
		return
	}
	for _, provider := range user1Providers {
		app.infoLog.Printf("Creating event from provider %s for user %d\n", provider.Name(), userID)

		token, err := app.models.AuthTokens.Token(userID, provider.Name())
		if err != nil {
			app.serverError(w, err)
			return
		}

		client := provider.CreateClient(r.Context(), token)
		err = provider.CreateEvent(userID, client, newEventData)
		if err != nil {
			app.errorLog.Fatalf("Error creating event: %v\n", err)
			return
		}
	}

	// Handle sending requests to user 2's providers (the user who created the request)
	user2Providers, err := providers.GetLinkedProviders(request.RequesterID, &app.models, app.googleOAuthConfig, app.azureOAuth2Config)
	if err != nil {
		app.serverError(w, err)
		return
	}
	for _, provider := range user2Providers {
		app.infoLog.Printf("Creating event from provider %s for user %d\n", provider.Name(), request.RequesterID)

		token, err := app.models.AuthTokens.Token(request.RequesterID, provider.Name())
		if err != nil {
			app.serverError(w, err)
			return
		}

		client := provider.CreateClient(r.Context(), token)
		err = provider.CreateEvent(request.RequesterID, client, newEventData)
		if err != nil {
			app.errorLog.Fatalf("Error creating event: %v\n", err)
			return
		}
	}

	// Save the new appointment to the database.

	err = app.models.AppointmentRequests.Delete(int(requestID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/requests", http.StatusSeeOther)
}
