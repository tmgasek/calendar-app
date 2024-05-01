package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
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

	// fully print the request
	fmt.Printf("request: %v\n", request)

	// Create events in the respective users' calendars using the request details
	// Here we need to send out the event to both users' linked calendars.
	// TODO: do this for all associated providers.
	// GOOGLE

	userTokenGoogle, err := app.models.AuthTokens.Token(userID, "google")
	if err != nil {
		app.serverError(w, err)
		return
	}
	targetUserTokenGoogle, err := app.models.AuthTokens.Token(int(request.RequesterID), "google")
	if err != nil {
		app.serverError(w, err)
		return
	}

	// fmt.Printf("startTime: %v\n", startTime)
	// fmt.Printf("endTime: %v\n", endTime)
	fmt.Printf("userTokenGoogle: %v\n", userTokenGoogle)
	fmt.Printf("targetUserTokenGoogle: %v\n", targetUserTokenGoogle)

	ctx := context.Background()
	userClient := app.googleOAuthConfig.Client(ctx, userTokenGoogle)
	targetUserClient := app.googleOAuthConfig.Client(ctx, targetUserTokenGoogle)

	userEvent := &calendar.Event{
		Summary:     request.Title,
		Description: request.Description,
		Start: &calendar.EventDateTime{
			DateTime: request.StartTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: request.StartTime.Format(time.RFC3339),
		},
	}

	user1GoogleService, err := calendar.NewService(ctx, option.WithHTTPClient(userClient))
	if err != nil {
		app.serverError(w, err)
		return
	}
	_, err = user1GoogleService.Events.Insert("primary", userEvent).Do()
	if err != nil {
		app.serverError(w, err)
		return
	}
	user2GoogleService, err := calendar.NewService(ctx, option.WithHTTPClient(targetUserClient))
	if err != nil {
		app.serverError(w, err)
		return
	}
	_, err = user2GoogleService.Events.Insert("primary", userEvent).Do()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// MICROSOFT
	outlookEvent := GraphEvent{
		Subject: request.Title,
		Body: struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		}{
			ContentType: "HTML",
			Content:     request.Description,
		},
		Start: struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		}{
			DateTime: request.StartTime.Format(time.RFC3339),
			TimeZone: "Pacific Standard Time", // or retrieve from user settings
		},
		End: struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		}{
			DateTime: request.EndTime.Format(time.RFC3339),
			TimeZone: "Pacific Standard Time",
		},
		Location: struct {
			DisplayName string `json:"displayName"`
		}{
			DisplayName: request.Location,
		},
	}

	// Send event to Microsoft for the user and target user
	user1MicrosoftToken, err := app.models.AuthTokens.Token(userID, "microsoft")
	user1azureClient := app.azureOAuth2Config.Client(r.Context(), user1MicrosoftToken)
	if err := createOutlookEvent(*user1azureClient, user1MicrosoftToken.AccessToken, outlookEvent); err != nil {
		app.serverError(w, err)
		return
	}

	user2MicrosoftToken, err := app.models.AuthTokens.Token(int(request.RequesterID), "microsoft")
	user2azureClient := app.azureOAuth2Config.Client(r.Context(), user2MicrosoftToken)
	if err := createOutlookEvent(*user2azureClient, user2MicrosoftToken.AccessToken, outlookEvent); err != nil {
		app.serverError(w, err)
		return
	}

	// Delete the appointment request from the database
	err = app.models.AppointmentRequests.Delete(int(requestID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/requests", http.StatusSeeOther)
}
