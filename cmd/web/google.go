package main

import (
	"context"
	"net/http"

	"github.com/tmgasek/calendar-app/internal/calendar"
	"golang.org/x/oauth2"
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

	client := app.googleOAuthConfig.Client(ctx, token)
	calendar.FetchAndSaveGoogleEvents(userID, client, &app.models)

	// Redirect back to homepage.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}