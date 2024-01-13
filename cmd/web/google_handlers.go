package main

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
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
	app.models.GoogleTokens.SaveToken(userID, token)
}

func (app *application) showEvents(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userID == 0 {
		// The user is not logged in, so redirect them to the login page.
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Retrieve the token from the database
	token, err := app.models.GoogleTokens.Token(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Refresh the token if it's expired
	if app.models.GoogleTokens.Expired(token) {
		app.infoLog.Println("*** Token expired, refreshing")
		refreshedToken, err := app.models.GoogleTokens.RefreshGoogleToken(userID, app.googleOAuthConfig, token)
		if err != nil {
			app.serverError(w, err)
			return
		}
		app.infoLog.Println("*** Refreshed token: ", refreshedToken)
		token = refreshedToken
	}

	// Use the token to create a Google Calendar service client
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
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}

			app.infoLog.Printf("%v (%v)\n", item.Summary, date)
		}
	}
}
