package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/tmgasek/calendar-app/internal/data"
	"golang.org/x/oauth2"
)

type MicrosoftCalendarProvider struct {
	token  *oauth2.Token
	config *oauth2.Config
}

func (p *MicrosoftCalendarProvider) Name() string {
	return "microsoft"
}

func (p *MicrosoftCalendarProvider) CreateClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return p.config.Client(ctx, token)
}

func (p *MicrosoftCalendarProvider) FetchEvents(userID int, client *http.Client) ([]data.Event, error) {
	// Define the time range for calendar events
	//TODO: need to handle cases if one of these dates is in different timezome
	// for example, endTime is now in british summer time.
	startTime := time.Now().Format("2006-01-02T15:04:05-07:00")
	// Make endtime one year from now
	endTime := time.Now().AddDate(1, 0, 0).Format("2006-01-02T15:04:05-07:00")

	// Create request to Microsoft Graph API
	reqURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/calendarview?startDateTime=%s&endDateTime=%s", url.QueryEscape(startTime), url.QueryEscape(endTime))

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	// Set the Authorization header with the access token
	req.Header.Set("Authorization", "Bearer "+p.token.AccessToken)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Read and log the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response body into the GraphEvent slice
	var resData struct {
		Value []GraphEvent `json:"value"`
	}

	if err := json.Unmarshal(body, &resData); err != nil {
		return nil, err
	}

	dbEvents := make([]data.Event, 0, len(resData.Value))
	for _, graphEvent := range resData.Value {
		event := convertGraphEventToEvent(userID, graphEvent)
		dbEvents = append(dbEvents, *event)
	}

	return dbEvents, nil
}

func (p *MicrosoftCalendarProvider) CreateEvent(userID int, client *http.Client, event data.Event) error {
	// Create event in Microsoft Calendar API
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
