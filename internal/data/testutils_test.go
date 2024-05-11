package data

import (
	"database/sql"
	"os"
	"os/exec"
	"testing"
)

func newTestDB(t *testing.T) *sql.DB {
	// Establish a sql.DB connection pool for our test database. Because our
	// setup and teardown scripts contains multiple SQL statements, we need
	// to use the "multiStatements=true" parameter in our DSN. This instructs
	// our MySQL database driver to support executing multiple SQL statements
	// in one db.Exec() call.
	testDSN := "postgres://test_calendar_app:password@localhost/test_calendar_app?sslmode=disable"

	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("migrate", "-path", "../../migrations", "-database", testDSN, "up")
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Seed the db.
	script, err := os.ReadFile("./testdata/seed.sql")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	// Use the t.Cleanup() to register a function *which will automatically be
	// called by Go when the current test (or sub-test) which calls newTestDB()
	// has finished*.
	t.Cleanup(func() {
		// Run the "down" migration to remove the test database schema.
		cmd := exec.Command("migrate", "-path", "../../migrations", "-database", testDSN, "down", "-all")
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			t.Fatal(err)
		}

		db.Close()
	})
	// Return the database connection pool.
	return db
}
