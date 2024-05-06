package data

import (
	"database/sql"
	"time"
)

type Appointment struct {
	ID              int
	CreatorID       int
	TargetID        int
	Title           string
	Description     string
	StartTime       time.Time
	EndTime         time.Time
	Location        string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	TimeZone        string
	Visibility      string
	Recurrence      string
	AppointmentType string
	GroupID         int
}

type AppointmentModel struct {
	DB *sql.DB
}

func (m *AppointmentModel) Insert(a *Appointment) (int, error) {
	var id int64
	query := `
		INSERT INTO appointments (creator_id, target_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone, visibility, recurrence, appointment_type, group_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`
	// We use a pointer here so that value can be null.
	var groupID *int
	if a.AppointmentType == "group" && a.GroupID != 0 {
		groupID = &a.GroupID
	}

	err := m.DB.QueryRow(query, a.CreatorID, a.TargetID, a.Title, a.Description, a.StartTime, a.EndTime, a.Location, a.Status, a.CreatedAt, a.UpdatedAt, a.TimeZone, a.Visibility, a.Recurrence, a.AppointmentType, groupID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *AppointmentModel) GetForUser(userID int) ([]*Appointment, error) {
	query := `
		SELECT id, creator_id, target_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone, visibility, recurrence, appointment_type, group_id
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
		var groupID sql.NullInt64

		err := rows.Scan(&a.ID, &a.CreatorID, &a.TargetID, &a.Title, &a.Description, &a.StartTime, &a.EndTime, &a.Location, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.TimeZone, &a.Visibility, &a.Recurrence, &a.AppointmentType, &groupID)
		if err != nil {
			return nil, err
		}

		if groupID.Valid {
			a.GroupID = int(groupID.Int64)
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
		SELECT id, creator_id, target_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone, visibility, recurrence, appointment_type, group_id
		FROM appointments
		WHERE id = $1
	`

	a := &Appointment{}
	var groupID sql.NullInt64

	err := m.DB.QueryRow(query, id).Scan(&a.ID, &a.CreatorID, &a.TargetID, &a.Title, &a.Description, &a.StartTime, &a.EndTime, &a.Location, &a.Status, &a.CreatedAt, &a.UpdatedAt, &a.TimeZone, &a.Visibility, &a.Recurrence, &a.AppointmentType, &groupID)
	if err != nil {
		return nil, err
	}

	if groupID.Valid {
		a.GroupID = int(groupID.Int64)
	}

	return a, nil
}
