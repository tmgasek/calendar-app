package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
	"golang.org/x/oauth2"
)

func (app *application) redirectToMicrosoftLogin(w http.ResponseWriter, r *http.Request) {
	url := app.azureOAuth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *application) handleMicrosoftAuthCallback(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userID == 0 {
		// The user is not logged in, so redirect them to the login page.
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	ctx := context.Background()
	code := r.URL.Query().Get("code")

	token, err := app.azureOAuth2Config.Exchange(ctx, code)
	if err != nil {
		app.errorLog.Printf("MICROSOFT ERROR: %v\n", err)
		// Handle error
		return
	}

	// Save token to the database.
	err = app.models.AuthTokens.SaveToken(userID, "microsoft", token)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Redirect back to homepage.
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (app *application) getOutlookEvents(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userID == 0 {
		// The user is not logged in, so redirect them to the login page.
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Retrieve the token from the database
	token, err := app.models.AuthTokens.Token(userID, "microsoft")
	if err != nil {
		app.serverError(w, err)
		return
	}

	client := app.azureOAuth2Config.Client(ctx, token)

	// Define the time range for calendar events
	startTime := time.Now().Format(time.RFC3339)
	endTime := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339) // Next 30 days

	// Create request to Microsoft Graph API
	reqURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/calendarview?startDateTime=%s&endDateTime=%s", startTime, endTime)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		app.errorLog.Printf("Error creating request: %v\n", err)
		// Handle error
		return
	}

	// Set the Authorization header with the access token
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		app.errorLog.Printf("Error making request: %v\n", err)
		// Handle error
		return
	}
	defer resp.Body.Close()

	// Read and log the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.errorLog.Printf("Error reading response body: %v\n", err)
		// Handle error
		return
	}

	// Unmarshal the response body into the GraphEvent slice
	var data struct {
		Value []GraphEvent `json:"value"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		app.errorLog.Printf("Error unmarshalling response body: %v\n", err)
		return
	}

	// Convert the Graph API events to your Event struct and save them
	for _, graphEvent := range data.Value {
		event := convertGraphEventToEvent(userID, graphEvent)

		// Save event to the database.
		err := app.models.Events.Insert(event)
		if err != nil {
			app.serverError(w, err)
			return
		}

		app.infoLog.Printf("Outlook event saved: %v (%v)\n", event.Title, event.StartTime)
	}

	// Log the events
	app.infoLog.Printf("Outlook Calendar Data: %s\n", string(body))
}

type GraphEvent struct {
	ID          string        `json:"id"`
	Subject     string        `json:"subject"`
	BodyPreview string        `json:"bodyPreview"`
	Start       GraphTime     `json:"start"`
	End         GraphTime     `json:"end"`
	Location    GraphLocation `json:"location"`
	IsAllDay    bool          `json:"isAllDay"`
}

type GraphTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type GraphLocation struct {
	DisplayName string `json:"displayName"`
}

func convertGraphEventToEvent(userID int, graphEvent GraphEvent) *data.Event {
	startTime, _ := time.Parse(time.RFC3339, graphEvent.Start.DateTime)
	endTime, _ := time.Parse(time.RFC3339, graphEvent.End.DateTime)

	return &data.Event{
		UserID:          userID,
		Provider:        "Microsoft",
		ProviderEventID: graphEvent.ID,
		Title:           graphEvent.Subject,
		Description:     graphEvent.BodyPreview,
		StartTime:       startTime,
		EndTime:         endTime,
		Location:        graphEvent.Location.DisplayName,
		IsAllDay:        graphEvent.IsAllDay,
		TimeZone:        graphEvent.Start.TimeZone,
	}
}
