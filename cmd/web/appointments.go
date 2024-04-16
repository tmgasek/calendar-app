package main

import (
	"context"
	"fmt"
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

	// Get the user and target users Google tokens
	// TODO: do this for all associated providers.
	userToken, err := app.models.AuthTokens.Token(userID, "google")
	if err != nil {
		app.serverError(w, err)
		return
	}
	targetUserToken, err := app.models.AuthTokens.Token(int(form.TargetUserID), "google")
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Printf("form: %v\n", form)
	fmt.Printf("startTime: %v\n", startTime)
	fmt.Printf("endTime: %v\n", endTime)
	fmt.Printf("userToken: %v\n", userToken)
	fmt.Printf("targetUserToken: %v\n", targetUserToken)

	ctx := context.Background()
	userClient := app.googleOAuthConfig.Client(ctx, userToken)
	targetUserClient := app.googleOAuthConfig.Client(ctx, targetUserToken)

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

	user1Service, err := calendar.NewService(ctx, option.WithHTTPClient(userClient))
	if err != nil {
		app.serverError(w, err)
		return
	}
	_, err = user1Service.Events.Insert("primary", userEvent).Do()
	if err != nil {
		app.serverError(w, err)
		return
	}

	user2Service, err := calendar.NewService(ctx, option.WithHTTPClient(targetUserClient))
	if err != nil {
		app.serverError(w, err)
		return
	}
	_, err = user2Service.Events.Insert("primary", userEvent).Do()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.infoLog.Println("Event created successfully!")
}
