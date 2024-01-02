package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var googleOauthConfig *oauth2.Config

func Setup() {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	googleOauthConfig = config
}

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

	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	// print out url
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *application) handleGoogleCalendarCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	code := r.URL.Query().Get("code")
	token, err := googleOauthConfig.Exchange(ctx, code)
	if err != nil {
		// Handle error.
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	app.infoLog.Println("Token: ", token)

	// tokenStore[userID] = token

	client := googleOauthConfig.Client(ctx, token)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		// Handle error.
		http.Error(w, "Failed to create calendar service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Fprintf(w, "Event: %s, Start Time: %s\n", item.Summary, item.Start.DateTime)
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}

	// Write the events to the response
	for _, item := range events.Items {
		fmt.Fprintf(w, "Event: %s, Start Time: %s\n", item.Summary, item.Start.DateTime)
	}
}
