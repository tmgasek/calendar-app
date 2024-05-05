package data

import (
	"database/sql"
	"time"
)

type Appointment struct {
	ID               int
	UserID           int
	MicrosoftEventID string
	GoogleEventID    string
	Title            string
	Description      string
	StartTime        time.Time
	EndTime          time.Time
	Location         string
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	TimeZone         string
	Visibility       string
	Recurrence       string
}

type AppointmentModel struct {
	DB *sql.DB
}

// Upserts the event.
func (m *AppointmentModel) Insert(a *Appointment) error {
	query := `
        INSERT INTO appointments (user_id, google_event_id, microsoft_event_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone, visibility, recurrence)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
    `

	_, err := m.DB.Exec(query, a.UserID, a.GoogleEventID, a.MicrosoftEventID, a.Title, a.Description, a.StartTime, a.EndTime, a.Location, a.Status, a.CreatedAt, a.UpdatedAt, a.TimeZone, a.Visibility, a.Recurrence)
	if err != nil {
		return err
	}

	return nil
}

func (m *AppointmentModel) GetByUserID(userID int) ([]*Appointment, error) {
	query := `
        SELECT id, user_id, google_event_id, microsoft_event_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone, visibility, recurrence
        FROM appointments
        WHERE user_id = $1
    `

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appointments := []*Appointment{}

	for rows.Next() {
		a := &Appointment{}
		err := rows.Scan(&a.ID, &a.UserID, &a.GoogleEventID, &a.MicrosoftEventID, &a.Title, &a.Description, &a.StartTime, &a.EndTime, &a.Location, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.TimeZone, &a.Visibility, &a.Recurrence)
		if err != nil {
			return nil, err
		}

		appointments = append(appointments, a)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return appointments, nil
}

func (m *AppointmentModel) Delete(id int) error {
	query := `
		DELETE FROM appointments
		WHERE id = $1
	`

	_, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}
