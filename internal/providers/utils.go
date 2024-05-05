package providers

import (
	"context"
	"net/http"

	"github.com/tmgasek/calendar-app/internal/data"
	"golang.org/x/oauth2"
)

func GetLinkedProviders(userID int, db *data.Models, googleConfig, microsoftConfig *oauth2.Config) ([]CalendarProvider, error) {
	var providers []CalendarProvider

	// Check for Google token.
	googleToken, err := db.AuthTokens.Token(userID, "google")
	if err != nil {
		return nil, err
	} else if googleToken != nil {
		providers = append(providers, &GoogleCalendarProvider{config: googleConfig})
	}

	// Check for Microsoft token.
	microsoftToken, err := db.AuthTokens.Token(userID, "microsoft")
	if err != nil {
		return nil, err
	} else if microsoftToken != nil {
		providers = append(providers, &MicrosoftCalendarProvider{token: microsoftToken, config: microsoftConfig})
	}

	return providers, nil
}

func GetProviderByName(userID int, name string, db *data.Models, googleConfig, microsoftConfig *oauth2.Config) (CalendarProvider, error) {
	switch name {
	case "google":
		return &GoogleCalendarProvider{config: googleConfig}, nil
	case "microsoft":
		microsoftToken, err := db.AuthTokens.Token(userID, "microsoft")
		if err != nil {
			return nil, err
		}
		return &MicrosoftCalendarProvider{token: microsoftToken, config: microsoftConfig}, nil
	default:
		return nil, nil
	}
}

func GetClient(provider CalendarProvider, userID int, db *data.Models) (*http.Client, error) {
	token, err := db.AuthTokens.Token(userID, provider.Name())
	if err != nil {
		return nil, err
	}
	client := provider.CreateClient(context.Background(), token)
	return client, nil
}
