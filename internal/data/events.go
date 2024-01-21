package data

import (
	"database/sql"
	"fmt"
	"time"
)

type Event struct {
	ID              int
	UserID          int
	Title           string
	Description     string
	StartDateTime   time.Time
	EndDateTime     time.Time
	CreatedDateTime time.Time
	UpdatedDateTime time.Time
}

type EventModel struct {
	DB *sql.DB
}

func (m *EventModel) Insert(event *Event) error {
	fmt.Println("INSERTING EVENT")
	fmt.Println(event)
	return nil
}

func (m *EventModel) GetByUserID(userID int) ([]*Event, error) {
	return nil, nil
}
