package calendar

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func FetchAndSaveGoogleEvents(userID int, client *http.Client, db *data.Models) error {
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	// Call the Google Calendar API to fetch events
	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		return err
	}

	// No events found
	if len(events.Items) == 0 {
		return nil
	}

	// Go over each event and save to db.
	for _, item := range events.Items {
		// Convert from Google event to own unified Event struct.
		event := convertGoogleEventToEvent(userID, item)

		// Save event to the database.
		err := db.Events.Insert(event)
		if err != nil {
			return err
		}

	}
	return nil
}

func convertGoogleEventToEvent(userID int, googleEvent *calendar.Event) *data.Event {
	event := &data.Event{
		UserID:          userID,
		Provider:        "Google",
		ProviderEventID: googleEvent.Id,
		Title:           googleEvent.Summary,
		Description:     googleEvent.Description,
		StartTime:       parseTime(googleEvent.Start.DateTime, googleEvent.Start.Date),
		EndTime:         parseTime(googleEvent.End.DateTime, googleEvent.End.Date),
		Location:        googleEvent.Location,
		IsAllDay:        googleEvent.Start.Date != "",
		Status:          googleEvent.Status,
		CreatedAt:       parseRFC3339Time(googleEvent.Created),
		UpdatedAt:       parseRFC3339Time(googleEvent.Updated),
		TimeZone:        googleEvent.Start.TimeZone, // Assuming Start and End TimeZones are the same
		Visibility:      googleEvent.Visibility,
		Recurrence:      strings.Join(googleEvent.Recurrence, ","),
	}

	return event
}

// Helper function to parse time in RFC3339 format
func parseRFC3339Time(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

// Helper function to parse Google event date and dateTime
func parseTime(dateTime, date string) time.Time {
	if dateTime != "" {
		return parseRFC3339Time(dateTime)
	}
	if date != "" {
		// Assuming date is in "YYYY-MM-DD" format
		t, _ := time.Parse("2006-01-02", date)
		return t
	}
	return time.Time{}
}
