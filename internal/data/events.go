package data

import (
	"database/sql"
	"time"
)

type Event struct {
	ID              int
	UserID          int
	Provider        string
	ProviderEventID string
	Title           string
	Description     string
	StartTime       time.Time
	EndTime         time.Time
	Location        string
	IsAllDay        bool
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	TimeZone        string
	Visibility      string
	Recurrence      string
}

type EventModel struct {
	DB *sql.DB
}

// Upserts the event.
func (m *EventModel) Insert(event *Event) error {
	query := `
		INSERT INTO events (user_id, provider, provider_event_id, title, description, start_time, end_time, location, is_all_day, status, created_at, updated_at, time_zone, visibility, recurrence)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (provider_event_id)
		DO UPDATE SET 
			title = EXCLUDED.title, 
			description = EXCLUDED.description, 
			start_time = EXCLUDED.start_time, 
			end_time = EXCLUDED.end_time, 
			location = EXCLUDED.location, 
			is_all_day = EXCLUDED.is_all_day, 
			status = EXCLUDED.status, 
			updated_at = EXCLUDED.updated_at, 
			time_zone = EXCLUDED.time_zone, 
			visibility = EXCLUDED.visibility, 
			recurrence = EXCLUDED.recurrence;
    `
	_, err := m.DB.Exec(query, event.UserID, event.Provider, event.ProviderEventID, event.Title, event.Description, event.StartTime, event.EndTime, event.Location, event.IsAllDay, event.Status, event.CreatedAt, event.UpdatedAt, event.TimeZone, event.Visibility, event.Recurrence)

	if err != nil {
		return err
	}
	return nil
}

func (m *EventModel) GetByUserID(userID int) ([]*Event, error) {
	query := `
        SELECT event_id, user_id, provider, provider_event_id, title, description, start_time, end_time, location, is_all_day, status, created_at, updated_at, time_zone, visibility, recurrence
        FROM events
        WHERE user_id = $1
    `

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []*Event{}

	for rows.Next() {
		event := &Event{}
		err := rows.Scan(&event.ID, &event.UserID, &event.Provider, &event.ProviderEventID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.IsAllDay, &event.Status, &event.CreatedAt, &event.UpdatedAt, &event.TimeZone, &event.Visibility, &event.Recurrence)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}