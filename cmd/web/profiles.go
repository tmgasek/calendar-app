package main

import (
	"net/http"
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
	"github.com/tmgasek/calendar-app/internal/providers"
)

type HourlyAvailability struct {
	Date  string // "2006-01-02" format
	Hours [24]string
}

func (app *application) userProfile(w http.ResponseWriter, r *http.Request) {
	templateData := app.newTemplateData(r)
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	allEvents, err := app.fetchEventsForUser(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	start := time.Now()
	end := start.AddDate(0, 0, 14)
	availability := app.initHourlyAvailability(start, end, allEvents)

	templateData.Events = allEvents
	templateData.HourlyAvailability = availability
	templateData.Hours = [16]int{7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22}

	app.render(w, http.StatusOK, "profile.tmpl", templateData)
}

func (app *application) viewUserProfile(w http.ResponseWriter, r *http.Request) {
	templateData := app.newTemplateData(r)

	currUserID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	targetUserID, err := app.readIDParam(r)
	if err != nil {
		app.notFound(w)
		return
	}

	if currUserID == int(targetUserID) {
		http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
		return
	}

	allEvents, err := app.fetchEventsForUser(int(targetUserID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	start := time.Now()
	end := start.AddDate(0, 0, 14)
	availability := app.initHourlyAvailability(start, end, allEvents)

	templateData.Events = allEvents
	templateData.HourlyAvailability = availability
	templateData.Hours = [16]int{7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22}
	templateData.TargetUserID = int(targetUserID)

	app.render(w, http.StatusOK, "user-calendar.tmpl", templateData)
}

func (app *application) fetchEventsForUser(userID int) ([]*data.Event, error) {
	var allEvents []*data.Event
	linkedProviders, err := providers.GetLinkedProviders(userID, &app.models, app.googleOAuthConfig, app.azureOAuth2Config)
	if err != nil {
		return nil, err
	}

	for _, p := range linkedProviders {
		app.infoLog.Printf("Getting events from provider %s for user %d\n", p.Name(), userID)
		client, err := providers.GetClient(p, userID, &app.models)
		if err != nil {
			return nil, err
		}

		events, err := p.FetchEvents(userID, client)
		if err != nil {
			return nil, err
		}

		for _, event := range events {
			// Make copy to avoid overwriting.
			eventCopy := event
			allEvents = append(allEvents, &eventCopy)
		}
	}

	return allEvents, nil
}

func (app *application) initHourlyAvailability(start, end time.Time, allEvents []*data.Event) []HourlyAvailability {
	availability := make([]HourlyAvailability, 0)
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		day := HourlyAvailability{
			Date:  d.Format("2006-01-02"),
			Hours: [24]string{},
		}
		for i := range day.Hours {
			day.Hours[i] = "free" // Init all to free
		}
		availability = append(availability, day)
	}

	for _, event := range allEvents {
		eventStart := event.StartTime
		eventEnd := event.EndTime
		if eventStart.Before(start) || eventEnd.After(end) {
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

	return availability
}
