package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robwestbrook/greenlight/internal/data"
	"github.com/robwestbrook/greenlight/internal/jsonlog"
	"github.com/robwestbrook/greenlight/internal/mailer"
)

// Declare a string containing the app version.
const version = "1.0.0"

// Define a config struct to hold all configuration
// settings fot the app. These settings will be
// read from command-line flags when the app starts.
// Configuration Settings:
//  1. port - network port for server
//  2. env - current operating environment
//	3. db - database config settings
//		a.	dsn - database name
//		b.	maxOpenConns - upper limit on open connections
//		c.	maxIdleConns - upper limit on idle connections
//		d.	maxIdleTime - set max time connection can be idle before expires
//	4. limiter - rate limiter config settings
//		a.	rps - requests per second
//		b.	burst - burst values
//		c.	enabled - enable/disable race limiting
//	5. smtp - email mailer config settings
//		a.	host - URL address for mailing host
//		b.	port - Host SMTP port
//		c.	username - username on host
//		d.	password - password on host
//		e.	sender - sender info used on host
//	6. CORS - CORS config settings
//		a.	trustedOrigins - slice containing trusted origins
type config struct {
	port int
	env  string
	db	 struct {
		dsn 					string
		maxOpenConns	int
		maxIdleConns	int
		maxIdleTime		string
	}
	limiter struct {
		rps			float64
		burst		int
		enabled	bool
	}
	smtp struct {
		host 			string
		port			int
		username	string
		password	string
		sender		string
	}
	cors struct {
		trustedOrigins []string
	}
}

// Define an app struct to hold dependencies.
// Dependencies:
//  1. config - the config struct
//  2. logger - System logger
//	3. models - the models struct
// 	4. mailer - the mailer struct
//	5. wg - wait group for goroutine monitoring
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg sync.WaitGroup
}

// main function - The entry point for the app.
func main() {
	// Load and read .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Fatal("Error converting PORT env to integer")
	}
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpSender := os.Getenv("SMTP_SENDER")

	// Declare an instance of config struct
	var cfg config

	// Read command-line flags
	// Flags:
	//	1.	Server port (default: 4000)
	//	2.	Environment setting (default: development)
	//	3.	Database name (default: greenlight.db)
	// 	4.	Max open DB connections (default: 25)
	//	5.	Max idle DB connections (default: 25)
	//	6.	Max DB idle time (default: 15 minutes)
	//	7.	Rate limiter requests per second (default: 2)
	//	8.	Rate Limiter  bursts (default: 4)
	//	9.	Rate limiter enabled (default: true)
	// 10.	SMTP host (default: .env host)
	// 11.	SMTP port (default: .env port)
	// 12.	SMTP username (default: .env username)
	// 13.	SMTP password (default: .env password)
	// 14.	SMTP sender (default: .env sender)
	// 15.	CORS trusted origins (default: empty []string slice)
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "greenlight.db", "SQLite database name")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "SQLite max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "SQLite max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "SQLite max connection idle time")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.StringVar(&cfg.smtp.host, "smtp-host", smtpHost, "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", smtpPort, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", smtpUsername, "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", smtpPassword, "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", smtpSender, "SMTP sender")
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	// Initialize a new jsonlogger that writes any 
	// messages *at or above* the INFO severity level
	// to the standard out stream.
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Call openDB() function to create connection pool,
	// passing in the config struct. If error returns,
	// log it and exit app immediately.
	logger.PrintInfo("Opening database connection pool", nil)
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Defer db.Close() so connection pool closes before
	// the main() function exits
	defer db.Close()

	// Log message db is open
	logger.PrintInfo("database connection pool established", nil)

	// Declare an instance of the application struct.
	// Contains:
	//	1.	cfg struct
	//	2.	logger
	//	3.	models - initialize a Models struct
	//	4.	mailer - initialize a new Mailer instance
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(
			cfg.smtp.host,
			cfg.smtp.port,
			cfg.smtp.username,
			cfg.smtp.password,
			cfg.smtp.sender,
		),
	}

	// Declare a new servermux.
	mux := http.NewServeMux()

	// Health Check route
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// Call app.serve(), in server.go to start server.
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

// openDB() function returns an sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set maximun number of open (in-use + idle)
	// connections in the pool. Value less than or equal
	// to 0 means no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set maximum number of idle connections in the pool.
	// Value less than or equal to 0 means no limit.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Use time.ParseDuration() function to convert idle
	// timeout string to a time.Duration type.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set maximum idle timeout
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5 second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to
	// the database, passing in the context as a parameter.
	// If connection is not established within the 5 
	// second deadline, return an error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return sql.DB connection pool
	return db, nil
}