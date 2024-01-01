package data

import (
	"context"
	"database/sql"
	// "errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// var (
// 	ErrDuplicateEmail = errors.New("duplicate email")
// )

type UserModel struct {
	DB *sql.DB
}

type User struct {
	ID           int
	Name         string
	Email        string
	PasswordHash []byte
	Created      time.Time
}

func (m *UserModel) Insert(name, email, password string) error {
	// Create a bcrypt hash of the plain text password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO users (name, email, password_hash, activated)
		VALUES($1, $2, $3, $4)
	`

	args := []any{name, email, hashedPassword, false}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}
