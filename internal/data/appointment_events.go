package data

import (
	"database/sql"
)

type AppointmentEvent struct {
	ID              int
	AppointmentID   int
	UserID          int
	ProviderName    string
	ProviderEventID string
}

type AppointmentEventModel struct {
	DB *sql.DB
}

type AppointmentEventModelInterface interface {
	Insert(event *AppointmentEvent) error
	GetByAppointmentID(appointmentID int) ([]*AppointmentEvent, error)
}

func (m *AppointmentEventModel) Insert(event *AppointmentEvent) error {
	query := `
		INSERT INTO appointment_events (appointment_id, user_id, provider_name, provider_event_id)
		VALUES ($1, $2, $3, $4)
	`
	_, err := m.DB.Exec(query, event.AppointmentID, event.UserID, event.ProviderName, event.ProviderEventID)

	if err != nil {
		return err
	}
	return nil
}

func (m *AppointmentEventModel) GetByAppointmentID(appointmentID int) ([]*AppointmentEvent, error) {
	query := `
		SELECT id, appointment_id, user_id, provider_name, provider_event_id
		FROM appointment_events
		WHERE appointment_id = $1
	`
	rows, err := m.DB.Query(query, appointmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []*AppointmentEvent{}
	for rows.Next() {
		event := &AppointmentEvent{}
		err := rows.Scan(&event.ID, &event.AppointmentID, &event.UserID, &event.ProviderName, &event.ProviderEventID)
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
