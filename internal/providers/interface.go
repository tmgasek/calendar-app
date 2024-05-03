package providers

import (
	"context"
	"net/http"

	"github.com/tmgasek/calendar-app/internal/data"
	"golang.org/x/oauth2"
)

type CalendarProvider interface {
	CreateClient(ctx context.Context, token *oauth2.Token) *http.Client
	FetchEvents(userID int, client *http.Client) ([]data.Event, error)
	CreateEvent(userID int, client *http.Client, event data.Event) error
	Name() string
}
