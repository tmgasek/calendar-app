package data

import (
	"database/sql"
	"time"
)

type Appointment struct {
	ID          int
	CreatorID   int
	TargetID    int
	Title       string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TimeZone    string
	Visibility  string
	Recurrence  string
}

type AppointmentModel struct {
	DB *sql.DB
}

func (m *AppointmentModel) Insert(a *Appointment) (int, error) {
	var id int64
	query := `
		INSERT INTO appointments (creator_id, target_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone, visibility, recurrence)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	err := m.DB.QueryRow(query, a.CreatorID, a.TargetID, a.Title, a.Description, a.StartTime, a.EndTime, a.Location, a.Status, a.CreatedAt, a.UpdatedAt, a.TimeZone, a.Visibility, a.Recurrence).Scan(&id)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *AppointmentModel) GetForUser(userID int) ([]*Appointment, error) {
	query := `
		SELECT id, creator_id, target_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone, visibility, recurrence
		FROM appointments
		WHERE creator_id = $1 OR target_id = $1
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appointments := []*Appointment{}

	for rows.Next() {
		a := &Appointment{}
		err := rows.Scan(&a.ID, &a.CreatorID, &a.TargetID, &a.Title, &a.Description, &a.StartTime, &a.EndTime, &a.Location, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.TimeZone, &a.Visibility, &a.Recurrence)
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

func (m *AppointmentModel) Get(id int) (*Appointment, error) {
	query := `
		SELECT id, creator_id, target_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone, visibility, recurrence
		FROM appointments
		WHERE id = $1
	`

	a := &Appointment{}

	err := m.DB.QueryRow(query, id).Scan(&a.ID, &a.CreatorID, &a.TargetID, &a.Title, &a.Description, &a.StartTime, &a.EndTime, &a.Location, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.TimeZone, &a.Visibility, &a.Recurrence)
	if err != nil {
		return nil, err
	}

	return a, nil
}
