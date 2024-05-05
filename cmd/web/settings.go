package main

import (
	"net/http"

	"github.com/tmgasek/calendar-app/internal/data"
	"github.com/tmgasek/calendar-app/internal/providers"
)

func (app *application) viewSettings(w http.ResponseWriter, r *http.Request) {
	templateData := app.newTemplateData(r)

	// Get the user ID from the session.
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	// Get the user record from the database.
	user, err := app.models.Users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// If the user record doesn't exist, return a 404 Not Found response.
	if user == nil {
		app.notFound(w)
		return
	}

	// We want to show if user has linkd their Google or Microsoft account.
	linkedProviders, err := providers.GetLinkedProviders(userID, &app.models, app.googleOAuthConfig, app.azureOAuth2Config)
	if err != nil {
		app.serverError(w, err)
		return
	}

	settings := &data.Settings{}

	for _, p := range linkedProviders {
		switch p.Name() {
		case "google":
			settings.LinkedGoogle = true
		case "microsoft":
			settings.LinkedMicrosoft = true
		}
	}

	// If the user record exists, add it to the template data.
	templateData.User = user
	templateData.Settings = settings

	// Render the profile settings page.
	app.render(w, http.StatusOK, "settings.tmpl", templateData)
}
