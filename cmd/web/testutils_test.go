package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	// "os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/tmgasek/calendar-app/internal/data/mocks"
)

var csrfTokenRX = regexp.MustCompile(`<input type="hidden" name="csrf_token" value="(.+)" />`)

func extractCSRFToken(t *testing.T, body string) string {
	matches := csrfTokenRX.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no csrf token found in body")
	}

	return html.UnescapeString(string(matches[1]))
}

func newTestApplication(t *testing.T) *application {
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	return &application{
		errorLog:       log.New(io.Discard, "", 0),
		infoLog:        log.New(io.Discard, "", 0),
		models:         mocks.NewMockModels(),
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		mailer:         mocks.NewMockMailer(),
	}
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	ts.Client().Jar = jar

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, string) {
	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

func (app *application) mockAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.sessionManager.Put(r.Context(), "authenticatedUserID", 1)
		next.ServeHTTP(w, r)
	})
}

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
	fmt.Println("Connected to test database")
	cmd := exec.Command("migrate", "-path", "../../migrations", "-database", testDSN, "up")
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	err = cmd.Run()
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
