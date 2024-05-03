package providers

import (
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
