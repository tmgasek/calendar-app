package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"golang.org/x/oauth2"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// create wrapper around our NotFound() helper.
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

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
	router.Handler(http.MethodGet, "/auth/microsoft", protected.ThenFunc(redirectToMicrosoftLogin))
	router.Handler(http.MethodGet, "/auth/microsoft-callback", protected.ThenFunc(app.handleMicrosoftAuthCallback))
	router.Handler(http.MethodPost, "/auth/microsoft-callback", protected.ThenFunc(app.handleMicrosoftAuthCallback))

	router.Handler(http.MethodGet, "/user/events", protected.ThenFunc(app.showEvents))

	// Create a new middleware chain.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}

func redirectToMicrosoftLogin(w http.ResponseWriter, r *http.Request) {
	url := microsoftOauth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *application) handleMicrosoftAuthCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Println("MICROSOFT CALLBACK")
	ctx := context.Background()
	code := r.URL.Query().Get("code")
	app.infoLog.Printf("MICROSOFT CODE: %v\n", code)

	token, err := microsoftOauth2Config.Exchange(ctx, code)
	if err != nil {
		app.errorLog.Printf("MICROSOFT ERROR: %v\n", err)
		// Handle error
		return
	}

	app.infoLog.Printf("MICROSOFT Token: %v\n", token)

	// Use token to access user's Outlook Calendar...
}
