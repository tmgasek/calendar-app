package data

import (
	"database/sql"
	"time"
)

type AppointmentRequest struct {
	RequestID    int
	RequesterID  int
	TargetUserID int
	Title        string
	Description  string
	StartTime    time.Time
	EndTime      time.Time
	Location     string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	TimeZone     string
}

type AppointmentRequestModel struct {
	DB *sql.DB
}

// Upserts the appointment request.
func (m *AppointmentRequestModel) Insert(request *AppointmentRequest) error {
	query := `
        INSERT INTO appointment_requests (requester_id, target_user_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (request_id)
        DO UPDATE SET 
            title = EXCLUDED.title, 
            description = EXCLUDED.description, 
            start_time = EXCLUDED.start_time, 
            end_time = EXCLUDED.end_time, 
            location = EXCLUDED.location, 
            status = EXCLUDED.status, 
            updated_at = EXCLUDED.updated_at, 
            time_zone = EXCLUDED.time_zone;
    `
	_, err := m.DB.Exec(query, request.RequesterID, request.TargetUserID, request.Title, request.Description, request.StartTime, request.EndTime, request.Location, request.Status, request.CreatedAt, request.UpdatedAt, request.TimeZone)

	if err != nil {
		return err
	}
	return nil
}

func (m *AppointmentRequestModel) GetForUser(userID int) ([]*AppointmentRequest, error) {
	query := `
        SELECT request_id, requester_id, target_user_id, title, description, start_time, end_time, location, status, created_at, updated_at, time_zone
        FROM appointment_requests
        WHERE target_user_id = $1
    `

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []*AppointmentRequest{}

	for rows.Next() {
		request := &AppointmentRequest{}
		err := rows.Scan(&request.RequestID, &request.RequesterID, &request.TargetUserID, &request.Title, &request.Description, &request.StartTime, &request.EndTime, &request.Location, &request.Status, &request.CreatedAt, &request.UpdatedAt, &request.TimeZone)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}
