package providers

import (
	"bytes"
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

func (p *MicrosoftCalendarProvider) DeleteEvent(userID int, client *http.Client, provider, eventID string) error {
	if provider != "microsoft" {
		return fmt.Errorf("invalid provider")
	}

	req, err := http.NewRequest("DELETE", "https://graph.microsoft.com/v1.0/me/events/"+eventID, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+p.token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete event: %s", resp.Status)
	}

	return nil
}

func (p *MicrosoftCalendarProvider) CreateEvent(userID int, client *http.Client, newEventData NewEventData) (eventID string, err error) {
	event := CreateGraphEventPayload{
		Subject: newEventData.Title,
		Body: struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		}{
			ContentType: "HTML",
			Content:     newEventData.Description,
		},
		Start: struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		}{
			DateTime: newEventData.StartTime.Format(time.RFC3339),
			TimeZone: "Pacific Standard Time", // or retrieve from user settings
		},
		End: struct {
			DateTime string `json:"dateTime"`
			TimeZone string `json:"timeZone"`
		}{
			DateTime: newEventData.EndTime.Format(time.RFC3339),
			TimeZone: "Pacific Standard Time",
		},
		Location: struct {
			DisplayName string `json:"displayName"`
		}{
			DisplayName: newEventData.Location,
		},
	}

	// Send the event to Microsoft
	eventJSON, err := json.Marshal(event)
	if err != nil {
		fmt.Println("error marshalling event")
		return "", err
	}

	req, err := http.NewRequest("POST", "https://graph.microsoft.com/v1.0/me/events", bytes.NewBuffer(eventJSON))
	if err != nil {
		fmt.Println("error creating request")
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+p.token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error sending request")
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		fmt.Println("error creating event")
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create event: %s", responseBody)
	}

	// Get the event ID from the response
	var resData struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&resData); err != nil {
		fmt.Println("error decoding response")
		return "", err
	}

	return resData.ID, nil
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

type CreateGraphEventPayload struct {
	Subject string `json:"subject"`
	Body    struct {
		ContentType string `json:"contentType"`
		Content     string `json:"content"`
	} `json:"body"`
	Start struct {
		DateTime string `json:"dateTime"`
		TimeZone string `json:"timeZone"`
	} `json:"start"`
	End struct {
		DateTime string `json:"dateTime"`
		TimeZone string `json:"timeZone"`
	} `json:"end"`
	Location struct {
		DisplayName string `json:"displayName"`
	} `json:"location"`
}
