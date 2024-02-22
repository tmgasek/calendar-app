package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/tmgasek/calendar-app/internal/data"
)

func (app *application) linkGoogleAccount(w http.ResponseWriter, r *http.Request) {
	// Get the user from the request context.
	// Get the Google OAuth2 URL from the provider, then add some additional
	// query string parameters, and then redirect the user to that URL.
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userID == 0 {
		// The user is not logged in, so redirect them to the login page.
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	url := app.googleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	// print out url
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *application) handleGoogleCalendarCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userID == 0 {
		// The user is not logged in, so redirect them to the login page.
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := app.googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		// Handle error.
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	app.infoLog.Println("Token: ", token)

	// Save token to the database.
	err = app.models.AuthTokens.SaveToken(userID, "google", token)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Redirect back to homepage.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) showEvents(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userID == 0 {
		// The user is not logged in, so redirect them to the login page.
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Retrieve the token from the database
	token, err := app.models.AuthTokens.Token(userID, "google")
	if err != nil {
		app.serverError(w, err)
		return
	}

	//!!!!!!!!!!!!!!  Maybe i don't need to refresh the token manually???

	// // Refresh the token if it's expired
	// if app.models.GoogleTokens.Expired(token) {
	// 	app.infoLog.Println("*** Token expired, refreshing")
	// 	refreshedToken, err := app.models.GoogleTokens.RefreshGoogleToken(userID, app.googleOAuthConfig, token)
	// 	if err != nil {
	// 		app.serverError(w, err)
	// 		return
	// 	}
	// 	app.infoLog.Println("*** Refreshed token: ", refreshedToken)
	// 	token = refreshedToken
	// }

	// Use the token to create a Google Calendar service client
	// This auto refreshes the token if it's expired.
	client := app.googleOAuthConfig.Client(context.Background(), token)
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Call the Google Calendar API to fetch events
	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		app.serverError(w, err)
		return
	}

	if len(events.Items) == 0 {
		app.infoLog.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			// Convert from Google event to own unified Event struct.
			event := convertGoogleEventToEvent(userID, item)

			// Save event to the database.
			err := app.models.Events.Insert(event)
			if err != nil {
				app.serverError(w, err)
				return
			}

			app.infoLog.Printf("Event saved: %v (%v)\n", item.Summary, item.Start.DateTime)
		}
	}
}

func convertGoogleEventToEvent(userID int, googleEvent *calendar.Event) *data.Event {
	event := &data.Event{
		UserID:          userID,
		Provider:        "Google",
		ProviderEventID: googleEvent.Id,
		Title:           googleEvent.Summary,
		Description:     googleEvent.Description,
		StartTime:       parseTime(googleEvent.Start.DateTime, googleEvent.Start.Date),
		EndTime:         parseTime(googleEvent.End.DateTime, googleEvent.End.Date),
		Location:        googleEvent.Location,
		IsAllDay:        googleEvent.Start.Date != "",
		Status:          googleEvent.Status,
		CreatedAt:       parseRFC3339Time(googleEvent.Created),
		UpdatedAt:       parseRFC3339Time(googleEvent.Updated),
		TimeZone:        googleEvent.Start.TimeZone, // Assuming Start and End TimeZones are the same
		Visibility:      googleEvent.Visibility,
		Recurrence:      strings.Join(googleEvent.Recurrence, ","),
	}

	return event
}

// Helper function to parse time in RFC3339 format
func parseRFC3339Time(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

// Helper function to parse Google event date and dateTime
func parseTime(dateTime, date string) time.Time {
	if dateTime != "" {
		return parseRFC3339Time(dateTime)
	}
	if date != "" {
		// Assuming date is in "YYYY-MM-DD" format
		t, _ := time.Parse("2006-01-02", date)
		return t
	}
	return time.Time{}
}
