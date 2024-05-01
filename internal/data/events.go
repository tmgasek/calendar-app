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
	RequesterID     int
	Requester       *Requester
}

type Requester struct {
	ID    int
	Name  string
	Email string
}

type EventModel struct {
	DB *sql.DB
}

// Upserts the event.
func (m *EventModel) Insert(event *Event) error {
	query := `
        INSERT INTO events (user_id, provider, provider_event_id, requester_id, title, description, start_time, end_time, location, is_all_day, status, created_at, updated_at, time_zone, visibility, recurrence)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
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
	_, err := m.DB.Exec(query, event.UserID, event.Provider, event.ProviderEventID, event.RequesterID, event.Title, event.Description, event.StartTime, event.EndTime, event.Location, event.IsAllDay, event.Status, event.CreatedAt, event.UpdatedAt, event.TimeZone, event.Visibility, event.Recurrence)
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

func (m *EventModel) GetPending(userID int) ([]*Event, error) {
	query := `
        SELECT e.event_id, e.user_id, e.provider, e.provider_event_id, e.requester_id, e.title, e.description, e.start_time, e.end_time, e.location, e.is_all_day, e.status, e.created_at, e.updated_at, e.time_zone, e.visibility, e.recurrence,
               u.id AS requester_id, u.name AS requester_name, u.email AS requester_email
        FROM events e
        LEFT JOIN users u ON e.requester_id = u.id
        WHERE e.user_id = $1 AND e.status = 'pending'
    `
	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []*Event{}

	for rows.Next() {
		event := &Event{}
		requester := &Requester{}

		err := rows.Scan(
			&event.ID, &event.UserID, &event.Provider, &event.ProviderEventID, &event.RequesterID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.IsAllDay, &event.Status, &event.CreatedAt, &event.UpdatedAt, &event.TimeZone, &event.Visibility, &event.Recurrence,
			&requester.ID, &requester.Name, &requester.Email,
		)
		if err != nil {
			return nil, err
		}

		event.Requester = requester
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (m *EventModel) UpdateStatus(eventID int, status string) error {
	query := `
		UPDATE events
		SET status = $1
		WHERE event_id = $2
	`

	_, err := m.DB.Exec(query, status, eventID)
	if err != nil {
		return err
	}

	return nil
}
