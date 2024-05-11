package mocks

import "github.com/tmgasek/calendar-app/internal/data"

type UserModel struct{}

var mockUser1 = &data.User{
	ID:    1,
	Name:  "Alice",
	Email: "alice@example.com",
}
var mockUser2 = &data.User{
	ID:    2,
	Name:  "Bob",
	Email: "bob@example.com",
}

func (m *UserModel) Insert(name, email, password string) error {
	switch email {
	case "dupe@example.com":
		return data.ErrDuplicateEmail
	default:
		return nil
	}
}
func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == mockUser1.Email && password == "pa$$word" {
		return 1, nil
	}
	return 0, data.ErrInvalidCredentials
}
func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	case 2:
		return true, nil
	default:
		return false, nil
	}
}

func (m *UserModel) Get(id int) (*data.User, error) {
	switch id {
	case 1:
		return mockUser1, nil
	case 2:
		return mockUser2, nil
	default:
		return nil, data.ErrRecordNotFound
	}
}

func (m *UserModel) SearchUsers(query string) ([]*data.User, error) {
	return []*data.User{
		mockUser1,
		mockUser2,
	}, nil
}

func (m *UserModel) GetByEmail(email string) (*data.User, error) {
	if email == mockUser1.Email {
		return mockUser1, nil
	}
	if email == mockUser2.Email {
		return mockUser2, nil
	}
	return nil, data.ErrRecordNotFound
}
