package data

import (
	"database/sql"
	"time"
)

type Requester struct {
	Name  string
	Email string
}

type AppointmentRequest struct {
	RequestID       int
	RequesterID     int
	GroupID         int
	TargetUserID    int
	Title           string
	Description     string
	StartTime       time.Time
	EndTime         time.Time
	Location        string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	TimeZone        string
	Requester       *Requester
	AppointmentType string
}

type AppointmentRequestModel struct {
	DB *sql.DB
}

// Upserts the appointment request.
func (m *AppointmentRequestModel) Insert(request *AppointmentRequest) error {
	query := `
        INSERT INTO appointment_requests (requester_id, target_user_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone, group_id, appointment_type)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        ON CONFLICT (request_id)
        DO UPDATE SET 
            title = EXCLUDED.title, 
            description = EXCLUDED.description, 
            start_time = EXCLUDED.start_time, 
            end_time = EXCLUDED.end_time, 
            location = EXCLUDED.location, 
            status = EXCLUDED.status, 
            updated_at = EXCLUDED.updated_at, 
            time_zone = EXCLUDED.time_zone,
			group_id = EXCLUDED.group_id,
			appointment_type = EXCLUDED.appointment_type;
    `

	// We use a pointer here so that value can be null.
	var groupID *int
	if request.AppointmentType == "group" && request.GroupID != 0 {
		groupID = &request.GroupID
	}

	_, err := m.DB.Exec(query, request.RequesterID, request.TargetUserID, request.Title, request.Description, request.StartTime, request.EndTime, request.Location, request.Status, request.CreatedAt, request.UpdatedAt, request.TimeZone, groupID, request.AppointmentType)

	if err != nil {
		return err
	}
	return nil
}

func (m *AppointmentRequestModel) GetForUser(userID int) ([]*AppointmentRequest, error) {
	query := `
        SELECT ar.request_id, ar.requester_id, ar.target_user_id, ar.title, ar.description, ar.start_time, ar.end_time, ar.location, ar.status, ar.created_at, ar.updated_at, ar.time_zone, ar.group_id, ar.appointment_type, u.name, u.email
        FROM appointment_requests ar
        JOIN users u ON ar.requester_id = u.id
        WHERE ar.target_user_id = $1
    `

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []*AppointmentRequest{}

	for rows.Next() {
		r := &AppointmentRequest{}
		var requesterName, requesterEmail string
		var groupID sql.NullInt64

		err := rows.Scan(&r.RequestID, &r.RequesterID, &r.TargetUserID, &r.Title, &r.Description, &r.StartTime, &r.EndTime, &r.Location, &r.Status, &r.CreatedAt, &r.UpdatedAt, &r.TimeZone, &groupID, &r.AppointmentType, &requesterName, &requesterEmail)
		if err != nil {
			return nil, err
		}

		if groupID.Valid {
			r.GroupID = int(groupID.Int64)
		}

		r.Requester = &Requester{
			Name:  requesterName,
			Email: requesterEmail,
		}

		requests = append(requests, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func (m *AppointmentRequestModel) Get(requestID int) (*AppointmentRequest, error) {
	query := `
		SELECT ar.request_id, ar.requester_id, ar.target_user_id, ar.title, ar.description, ar.start_time, ar.end_time, ar.location, ar.status, ar.created_at, ar.updated_at, ar.time_zone, ar.group_id, ar.appointment_type, u.name, u.email
		FROM appointment_requests ar
		JOIN users u ON ar.requester_id = u.id
		WHERE ar.request_id = $1
	`

	row := m.DB.QueryRow(query, requestID)

	r := &AppointmentRequest{}
	var requesterName, requesterEmail string
	var groupID sql.NullInt64

	err := row.Scan(&r.RequestID, &r.RequesterID, &r.TargetUserID, &r.Title, &r.Description, &r.StartTime, &r.EndTime, &r.Location, &r.Status, &r.CreatedAt, &r.UpdatedAt, &r.TimeZone, &groupID, &r.AppointmentType, &requesterName, &requesterEmail)
	if err != nil {
		return nil, err
	}

	if groupID.Valid {
		r.GroupID = int(groupID.Int64)
	}

	r.Requester = &Requester{
		Name:  requesterName,
		Email: requesterEmail,
	}

	return r, nil
}

func (m *AppointmentRequestModel) Delete(requestID int) error {
	query := `
		DELETE FROM appointment_requests WHERE request_id = $1
	`
	_, err := m.DB.Exec(query, requestID)
	if err != nil {
		return err
	}
	return nil
}
