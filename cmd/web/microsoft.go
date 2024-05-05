package main

import (
	"context"
	"net/http"

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

	app.sessionManager.Put(r.Context(), "flash", "Microsoft account linked successfully!")
	// Redirect back to homepage.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
