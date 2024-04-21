package calendar

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
)

func FetchAndSaveOutlookEvents(userID int, client *http.Client, db *data.Models) error {
	token, err := db.AuthTokens.Token(userID, "microsoft")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No token found, so return without error
			return nil
		}
		return err
	}

	// Define the time range for calendar events
	//TODO: need to handle cases if one of these dates is in different timezome
	// for example, endTime is now in british summer time.
	startTime := time.Now().Format("2006-01-02T15:04:05-07:00")
	// Make endtime one year from now
	endTime := time.Now().AddDate(1, 0, 0).Format("2006-01-02T15:04:05-07:00")

	fmt.Printf("startTime: %v\n", startTime)
	fmt.Printf("endTime: %v\n", endTime)

	// Create request to Microsoft Graph API
	reqURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/calendarview?startDateTime=%s&endDateTime=%s", url.QueryEscape(startTime), url.QueryEscape(endTime))

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}

	// Set the Authorization header with the access token
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Read and log the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Unmarshal the response body into the GraphEvent slice
	var data struct {
		Value []GraphEvent `json:"value"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	// Convert the Graph API events to your Event struct and save them
	for _, graphEvent := range data.Value {
		event := convertGraphEventToEvent(userID, graphEvent)

		// Save event to the database.
		err := db.Events.Insert(event)
		if err != nil {
			return err
		}
	}

	return nil
}

type GraphEvent struct {
	ID                   string        `json:"id"`
	Subject              string        `json:"subject"`
	BodyPreview          string        `json:"bodyPreview"`
	Start                GraphTime     `json:"start"`
	End                  GraphTime     `json:"end"`
	Location             GraphLocation `json:"location"`
	IsAllDay             bool          `json:"isAllDay"`
	CreatedDateTime      string        `json:"createdDateTime"`
	LastModifiedDateTime string        `json:"lastModifiedDateTime"`
}

type GraphTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type GraphLocation struct {
	DisplayName string `json:"displayName"`
}

func convertGraphEventToEvent(userID int, graphEvent GraphEvent) *data.Event {
	startTime, _ := time.Parse("2006-01-02T15:04:05.0000000", graphEvent.Start.DateTime)
	endTime, _ := time.Parse("2006-01-02T15:04:05.0000000", graphEvent.End.DateTime)

	createdAt, _ := time.Parse(time.RFC3339, graphEvent.CreatedDateTime)
	updatedAt, _ := time.Parse(time.RFC3339, graphEvent.LastModifiedDateTime)

	return &data.Event{
		UserID:          userID,
		Provider:        "Microsoft",
		ProviderEventID: graphEvent.ID,
		Title:           graphEvent.Subject,
		Description:     graphEvent.BodyPreview,
		StartTime:       startTime,
		EndTime:         endTime,
		Location:        graphEvent.Location.DisplayName,
		IsAllDay:        graphEvent.IsAllDay,
		TimeZone:        graphEvent.Start.TimeZone,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}
