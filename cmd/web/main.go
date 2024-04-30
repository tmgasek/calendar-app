package main

import (
	"context"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/form/v4"
	_ "github.com/lib/pq"
	"github.com/tmgasek/calendar-app/internal/data"
	"github.com/tmgasek/calendar-app/internal/mailer"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
	"google.golang.org/api/calendar/v3"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
)

// Struct to hold all config settings for the app.
// Will read in these settings from cmd flags.
type config struct {
	addr string
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

// App struct to hold the app-wide dependencies.
type application struct {
	errorLog          *log.Logger
	infoLog           *log.Logger
	templateCache     map[string]*template.Template
	formDecoder       *form.Decoder
	models            data.Models
	sessionManager    *scs.SessionManager
	googleOAuthConfig *oauth2.Config
	azureOAuth2Config *oauth2.Config
	mailer            mailer.Mailer
}

func main() {
	var cfg config

	// Load in .env file.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	flag.StringVar(&cfg.addr, "addr", ":8080", "HTTP network address")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// DB.
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "Postgresql DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	// Mailer.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "a7303712992c04", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "fcbce4d2ec04cd", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "calendar-genie <no-reply@calendar-genie>", "SMTP sender")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(cfg)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		models:         data.NewModels(db),
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username,
			cfg.smtp.password, cfg.smtp.sender),
	}

	srv := &http.Server{
		Addr:         cfg.addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	app.initGoogleAuthConfig()
	app.initAzureAuthConfig()

	infoLog.Printf("Starting server on %s", cfg.addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)

}

func openDB(cfg config) (*sql.DB, error) {
	// Create empty conn pool.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	// Create ctx with a 5 sec timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Establish a new conn to the db passing in ctx. If conn couldn't be
	// established successfully withing 5 secs, this will return err.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// initAzureAuthConfig initializes the oauth2 config for Azure and
// stores it in the app struct.
func (app *application) initAzureAuthConfig() {
	app.azureOAuth2Config = &oauth2.Config{
		ClientID:     os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("AZURE_REDIRECT_URL"),
		Scopes:       []string{"https://graph.microsoft.com/Calendars.ReadWrite", "offline_access"},
		Endpoint:     microsoft.AzureADEndpoint("common"),
	}
}

func (app *application) initGoogleAuthConfig() {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)

	// app.googleOAuthConfig = &oauth2.Config{
	// 	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	// 	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	// 	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	// 	Scopes:       []string{calendar.CalendarScope},
	// }

	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	app.googleOAuthConfig = config
}
