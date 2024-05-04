package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/tmgasek/calendar-app/internal/data"
	"github.com/tmgasek/calendar-app/internal/validator"
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

	targetUser, err := app.models.Users.Get(int(form.TargetUserID))
	if err != nil {
		app.serverError(w, err)
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

	requestee, err := app.models.Users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	type EmailData struct {
		RequesteeName string
	}

	// Create the appointment request.
	appointmentRequest := &data.AppointmentRequest{
		RequesterID:  int(userID),
		TargetUserID: int(form.TargetUserID),
		Title:        form.Title,
		Description:  form.Description,
		StartTime:    startTime,
		EndTime:      endTime,
		Location:     form.Location,
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = app.models.AppointmentRequests.Insert(appointmentRequest)
	if err != nil {
		app.errorLog.Println(err)
		app.serverError(w, err)
		return
	}

	emailData := EmailData{
		RequesteeName: requestee.Name,
	}

	err = app.mailer.Send(targetUser.Email, "confirm-appointment.tmpl", emailData)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.infoLog.Println("********** Email sent")
}

func (app *application) deleteAppointment(w http.ResponseWriter, r *http.Request) {
	// Get the event ID from the URL parameters.
	params := httprouter.ParamsFromContext(r.Context())
	eventID := params.ByName("id")
	if eventID == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Get the provider from the URL query parameters.
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	app.infoLog.Printf("Deleting event %s from provider %s for user %d\n", eventID, provider, userID)

	// Delete the event from the provider's calendar
	// err := app.deleteEventFromProvider(userID, provider, eventID)
	// if err != nil {
	// 	app.serverError(w, err)
	// 	return
	// }

	// Remove the event from the database
	// err = app.models.Events.DeleteByProviderEventID(eventID)
	// if err != nil {
	// 	app.serverError(w, err)
	// 	return
	// }

	// Redirect back to the profile page
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}
