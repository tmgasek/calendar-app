package mocks

import "github.com/tmgasek/calendar-app/internal/data"

type UserModel struct{}

var mockUser = &data.User{
	ID:    1,
	Name:  "Alice",
	Email: "alice@example.com",
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
	if email == mockUser.Email && password == "pa$$word" {
		return 1, nil
	}
	return 0, data.ErrInvalidCredentials
}
func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (m *UserModel) Get(id int) (*data.User, error) {
	switch id {
	case 1:
		return mockUser, nil
	default:
		return nil, data.ErrRecordNotFound
	}
}

func (m *UserModel) SearchUsers(query string) ([]*data.User, error) {
	return []*data.User{
		mockUser,
	}, nil
}

func (m *UserModel) GetByEmail(email string) (*data.User, error) {
	if email == mockUser.Email {
		return mockUser, nil
	}
	return nil, data.ErrRecordNotFound
}
