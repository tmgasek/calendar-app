package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type GoogleCalendarProvider struct {
	config *oauth2.Config
}

func (p *GoogleCalendarProvider) CreateClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return p.config.Client(ctx, token)
}

func (p *GoogleCalendarProvider) Name() string {
	return "google"
}

func (p *GoogleCalendarProvider) DeleteEvent(userID int, client *http.Client, provider, eventID string) error {
	if provider != "google" {
		return fmt.Errorf("invalid provider")
	}
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	err = srv.Events.Delete("primary", eventID).Do()
	if err != nil {
		return err
	}

	return nil
}

func (p *GoogleCalendarProvider) CreateEvent(userID int, client *http.Client, newEventData NewEventData) (eventID string, err error) {
	event := &calendar.Event{
		Summary:     newEventData.Title,
		Description: newEventData.Description,
		Start: &calendar.EventDateTime{
			DateTime: newEventData.StartTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: newEventData.EndTime.Format(time.RFC3339),
		},
	}

	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return "", err
	}

	googleEvent, err := srv.Events.Insert("primary", event).Do()
	if err != nil {
		return "", err
	}

	return googleEvent.Id, nil
}

func (p *GoogleCalendarProvider) FetchEvents(userID int, client *http.Client) ([]data.Event, error) {
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	// Call the Google Calendar API to fetch events
	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		return nil, err
	}

	// No events found
	if len(events.Items) == 0 {
		return nil, nil
	}

	dbEvents := make([]data.Event, 0, len(events.Items))

	// Go over each event and save to db.
	for _, item := range events.Items {
		// Convert from Google event to own unified Event struct.
		event := convertGoogleEventToEvent(userID, item)
		dbEvents = append(dbEvents, *event)
	}
	return dbEvents, nil
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
