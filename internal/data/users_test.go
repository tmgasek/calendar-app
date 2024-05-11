package data

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/tmgasek/calendar-app/internal/assert"
)

func TestUserModelExists(t *testing.T) {
	tests := []struct {
		name   string
		userID int
		want   bool
	}{
		{
			name:   "Valid ID",
			userID: 1,
			want:   true,
		},
		{
			name:   "Zero ID",
			userID: 0,
			want:   false,
		},
		{
			name:   "Non-existent ID",
			userID: 999,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the newTestDB() helper function to get a connection pool to
			// our test database. Calling this here -- inside t.Run() -- means
			// that fresh database tables and data will be set up and torn down
			// for each sub-test.
			db := newTestDB(t)
			// Create a new instance of the UserModel.
			m := UserModel{db}
			// Call the UserModel.Exists() method and check that the return
			// value and error match the expected values for the sub-test.
			exists, err := m.Exists(tt.userID)
			assert.Equal(t, exists, tt.want)
			assert.NilError(t, err)
		})
	}
}

func TestUserModelInsert(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		email    string
		password string
		wantErr  error
	}{
		{
			name:     "Valid user",
			userName: "John Doe",
			email:    "john@example.com",
			password: "password123",
			wantErr:  nil,
		},
		{
			name:     "Duplicate email",
			userName: "Jane Smith",
			email:    "alice@example.com",
			password: "password456",
			wantErr:  ErrDuplicateEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := UserModel{db}

			err := m.Insert(tt.userName, tt.email, tt.password)
			assert.Equal(t, err, tt.wantErr)
		})
	}
}

func TestUserModelAuthenticate(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantID   int
		wantErr  error
	}{
		{
			name:     "Valid credentials",
			email:    "alice@example.com",
			password: "pa$$word",
			wantID:   1,
			wantErr:  nil,
		},
		{
			name:     "Invalid email",
			email:    "invalid@example.com",
			password: "pa$$word",
			wantID:   0,
			wantErr:  ErrInvalidCredentials,
		},
		{
			name:     "Invalid password",
			email:    "alice@example.com",
			password: "invalidpassword",
			wantID:   0,
			wantErr:  ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := UserModel{db}

			id, err := m.Authenticate(tt.email, tt.password)
			assert.Equal(t, id, tt.wantID)
			assert.Equal(t, err, tt.wantErr)
		})
	}
}

func TestUserModelGet(t *testing.T) {
	tests := []struct {
		name    string
		userID  int
		want    *User
		wantErr error
	}{
		{
			name:   "Valid ID",
			userID: 1,
			want: &User{
				ID:    1,
				Name:  "Alice",
				Email: "alice@example.com",
			},
			wantErr: nil,
		},
		{
			name:    "Non-existent ID",
			userID:  999,
			want:    nil,
			wantErr: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := UserModel{db}

			user, err := m.Get(tt.userID)
			if user != nil {
				assert.Equal(t, user.ID, tt.want.ID)
				assert.Equal(t, user.Name, tt.want.Name)
				assert.Equal(t, user.Email, tt.want.Email)
			} else {
				assert.Equal(t, user, tt.want)
			}

			assert.Equal(t, err, tt.wantErr)
		})
	}
}

func TestUserModelSearchUsers(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantUsers []*User
		wantErr   error
	}{
		{
			name:  "Valid query",
			query: "alice",
			wantUsers: []*User{
				{
					ID:    1,
					Name:  "Alice",
					Email: "alice@example.com",
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := UserModel{db}

			users, err := m.SearchUsers(tt.query)

			if err != tt.wantErr {
				t.Errorf("got error %v; want error %v", err, tt.wantErr)
			}

			if len(users) != len(tt.wantUsers) {
				t.Errorf("got %d users; want %d users", len(users), len(tt.wantUsers))
			}

			for i, user := range users {
				if i < len(tt.wantUsers) {
					assert.Equal(t, user.ID, tt.wantUsers[i].ID)
					assert.Equal(t, user.Name, tt.wantUsers[i].Name)
					assert.Equal(t, user.Email, tt.wantUsers[i].Email)
				}
			}
		})
	}
}

func TestUserModelGetByEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		want    *User
		wantErr error
	}{
		{
			name:  "Valid email",
			email: "alice@example.com",
			want: &User{
				ID:    1,
				Name:  "Alice",
				Email: "alice@example.com",
			},
			wantErr: nil,
		},
		{
			name:    "Non-existent email",
			email:   "nonexistent@example.com",
			want:    nil,
			wantErr: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := UserModel{db}

			user, err := m.GetByEmail(tt.email)
			if user != nil {
				assert.Equal(t, user.ID, tt.want.ID)
				assert.Equal(t, user.Name, tt.want.Name)
				assert.Equal(t, user.Email, tt.want.Email)
			} else {
				assert.Equal(t, user, tt.want)
			}
			assert.Equal(t, err, tt.wantErr)
		})
	}
}
