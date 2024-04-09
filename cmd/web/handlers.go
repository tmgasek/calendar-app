package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
	"github.com/tmgasek/calendar-app/internal/validator"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, http.StatusOK, "home.tmpl", data)
}

// Easily render templates from the cache
func (app *application) render(
	w http.ResponseWriter,
	status int,
	page string,
	data *templateData,
) {
	// retrieve the right template set from cache based on page name
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("The template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	// init new buffer
	buf := new(bytes.Buffer)

	// write template to buffer instead of to the http.ResponseWriter
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// if template written to buffer w/o errors, we are good to go.
	w.WriteHeader(status)

	buf.WriteTo(w)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.tmpl", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form.
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	// Check if credentials are valid. If not, add generic non field err.
	id, err := app.models.Users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, data.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Change session ID. Good practice to do when auth / privilege state changes.
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Add ID of current user to session, so they are now "logged in".
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, http.StatusOK, "signup.tmpl", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	// 1. Check that the provided name, email address and password are not blank.
	// 2. Sanity check the format of the email address.
	// 3. Ensure that the password is at least 8 characters long.
	// 4. Make sure that the email address isn’t already in use.
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "title", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 chars long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	err = app.models.Users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, data.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address already in use!")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Signup successful! Please login.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "Log out successful!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type HourlyAvailability struct {
	Date  string // "2006-01-02" format
	Hours [24]string
}

func (app *application) userProfile(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	// Get the event data
	events, err := app.models.Events.GetByUserID(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Determine the range of dates to display. For now show 14 days from today.
	start := time.Now()
	end := start.AddDate(0, 0, 14)

	// Init hourly availability
	availability := make([]HourlyAvailability, 0)
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		day := HourlyAvailability{
			Date:  d.Format("2006-01-02"),
			Hours: [24]string{},
		}

		for i := range day.Hours {
			// Init all to free
			day.Hours[i] = "free"
		}
		availability = append(availability, day)
	}

	// Mark the hours that are busy based on user's events
	for _, event := range events {
		eventStart := event.StartTime
		eventEnd := event.EndTime

		// Only process event within our range
		if eventStart.Before(start) || eventEnd.After(end) {
			fmt.Printf("Event outside of range: %s - %s\n", eventStart, eventEnd)
			continue
		}

		for i := range availability {
			day := &availability[i]
			if eventStart.Format("2006-01-02") != day.Date {
				continue
			}

			startHour := eventStart.Hour()
			endHour := eventEnd.Hour()

			fmt.Printf("Start: %d, End: %d\n", startHour, endHour)

			for h := startHour; h <= endHour && h < 24; h++ {
				fmt.Printf("Marking %d as busy\n", h)
				day.Hours[h] = "busy"
			}
		}
	}

	fmt.Printf("availability: %v\n", availability)

	data.Events = events
	data.HourlyAvailability = availability
	data.Hours = [16]int{
		7, 8, 9, 10, 11,
		12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
	}

	app.render(w, http.StatusOK, "profile.tmpl", data)
}
