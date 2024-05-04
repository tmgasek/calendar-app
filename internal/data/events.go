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
