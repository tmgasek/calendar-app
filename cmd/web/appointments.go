package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tmgasek/calendar-app/internal/validator"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
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
	type EmailData struct {
		RequesteeName string
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
	return

	// Here we need to send out the event to both users' linked calendars.
	// TODO: do this for all associated providers.
	// GOOGLE

	userTokenGoogle, err := app.models.AuthTokens.Token(userID, "google")
	if err != nil {
		app.serverError(w, err)
		return
	}
	targetUserTokenGoogle, err := app.models.AuthTokens.Token(int(targetUser.ID), "google")
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Printf("form: %v\n", form)
	fmt.Printf("startTime: %v\n", startTime)
	fmt.Printf("endTime: %v\n", endTime)
	fmt.Printf("userTokenGoogle: %v\n", userTokenGoogle)
	fmt.Printf("targetUserTokenGoogle: %v\n", targetUserTokenGoogle)

	ctx := context.Background()
	userClient := app.googleOAuthConfig.Client(ctx, userTokenGoogle)
	targetUserClient := app.googleOAuthConfig.Client(ctx, targetUserTokenGoogle)

	userEvent := &calendar.Event{
		Summary:     form.Title,
		Description: form.Description,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
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
		Subject: form.Title,
		Body: struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		}{
			ContentType: "HTML",
			Content:     form.Description,
		},
		Start: struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		}{
			DateTime: startTime.Format(time.RFC3339),
			TimeZone: "Pacific Standard Time", // or retrieve from user settings
		},
		End: struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		}{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "Pacific Standard Time",
		},
		Location: struct {
			DisplayName string `json:"displayName"`
		}{
			DisplayName: form.Location,
		},
	}

	// Send event to Microsoft for the user and target user
	user1MicrosoftToken, err := app.models.AuthTokens.Token(userID, "microsoft")
	user1azureClient := app.azureOAuth2Config.Client(r.Context(), user1MicrosoftToken)
	if err := createOutlookEvent(*user1azureClient, user1MicrosoftToken.AccessToken, outlookEvent); err != nil {
		app.serverError(w, err)
		return
	}

	user2MicrosoftToken, err := app.models.AuthTokens.Token(int(targetUser.ID), "microsoft")
	user2azureClient := app.azureOAuth2Config.Client(r.Context(), user2MicrosoftToken)
	if err := createOutlookEvent(*user2azureClient, user2MicrosoftToken.AccessToken, outlookEvent); err != nil {
		app.serverError(w, err)
		return
	}

	// Redirect back to profile
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}

type GraphEvent struct {
	Subject string `json:"subject"`
	Body    struct {
		ContentType string `json:"contentType"`
		Content     string `json:"content"`
	} `json:"body"`
	Start struct {
		DateTime string `json:"dateTime"`
		TimeZone string `json:"timeZone"`
	} `json:"start"`
	End struct {
		DateTime string `json:"dateTime"`
		TimeZone string `json:"timeZone"`
	} `json:"end"`
	Location struct {
		DisplayName string `json:"displayName"`
	} `json:"location"`
}

func createOutlookEvent(client http.Client, userToken string, event GraphEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://graph.microsoft.com/v1.0/me/events", bytes.NewBuffer(eventJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create event: %s", responseBody)
	}
	return nil
}
