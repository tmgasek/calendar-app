package mocks

import "github.com/tmgasek/calendar-app/internal/data"

type GroupModel struct{}

var mockGroup = &data.Group{
	ID:          1,
	Name:        "Test Group",
	Description: "Test Description",
	CreatedAt:   "2021-01-01T00:00:00Z",
	UpdatedAt:   "2021-01-01T00:00:00Z",
	Members: []*data.GroupMember{
		{
			ID:    1,
			Name:  "Alice",
			Email: "alice@example.com",
		},
	},
}

func (m *GroupModel) Insert(userID int, name, description string) (int, error) {
	return 1, nil
}

func (m *GroupModel) Get(id int) (*data.Group, error) {
	switch id {
	case 1:
		return mockGroup, nil
	default:
		return nil, data.ErrRecordNotFound
	}
}

func (m *GroupModel) GetAllForUser(userID int) ([]*data.Group, error) {
	return []*data.Group{mockGroup}, nil
}

func (m *GroupModel) AddMember(groupID, userID int) error {
	return nil
}
