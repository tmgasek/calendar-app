package main

import (
	"fmt"
	"net/http"
	"time"
)

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

			for h := startHour; h <= endHour && h < 24; h++ {
				day.Hours[h] = "busy"
			}
		}
	}

	data.Events = events
	data.HourlyAvailability = availability
	data.Hours = [16]int{
		7, 8, 9, 10, 11,
		12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
	}

	app.render(w, http.StatusOK, "profile.tmpl", data)
}

func (app *application) viewUserProfile(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	fmt.Printf("userID: %v\n", userID)

	targetUserID, err := app.readIDParam(r)
	if err != nil {
		app.notFound(w)
		return
	}

	// Get the event data
	events, err := app.models.Events.GetByUserID(int(targetUserID))
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

			for h := startHour; h <= endHour && h < 24; h++ {
				day.Hours[h] = "busy"
			}
		}
	}

	data.Events = events
	data.HourlyAvailability = availability
	data.Hours = [16]int{
		7, 8, 9, 10, 11,
		12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
	}

	app.render(w, http.StatusOK, "user-calendar.tmpl", data)
}
