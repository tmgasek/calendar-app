package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/tmgasek/calendar-app/ui"
	"golang.org/x/oauth2"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// create wrapper around our NotFound() helper.
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	router.HandlerFunc(http.MethodGet, "/ping", ping)

	// For dynamic routes.
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	// Protected application routes.
	protected := dynamic.Append(app.requireAuthentication)
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))

	router.Handler(http.MethodGet, "/user/profile", protected.ThenFunc(app.userProfile))

	// Google OAuth routes.
	router.Handler(http.MethodGet, "/google/link", protected.ThenFunc(app.linkGoogleAccount))
	router.Handler(http.MethodGet, "/auth/callback", protected.ThenFunc(app.handleGoogleCalendarCallback))

	// Microsoft OAuth routes.
	router.Handler(http.MethodGet, "/auth/azure/link", protected.ThenFunc(app.redirectToMicrosoftLogin))
	router.Handler(http.MethodGet, "/auth/microsoft-callback", protected.ThenFunc(app.handleMicrosoftAuthCallback))

	router.Handler(http.MethodGet, "/user/events", protected.ThenFunc(app.showEvents))
	router.Handler(http.MethodGet, "/user/outlook/events", protected.ThenFunc(app.getOutlookEvents))

	// Create a new middleware chain.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}

func (app *application) redirectToMicrosoftLogin(w http.ResponseWriter, r *http.Request) {
	// log the config
	fmt.Println("MICROSOFT CONFIG", app.azureOAuth2Config)

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
	err = app.models.GoogleTokens.SaveToken(userID, token)
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
	token, err := app.models.GoogleTokens.Token(userID)
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

	// Log the events
	app.infoLog.Printf("Outlook Calendar Data: %s\n", string(body))
}
