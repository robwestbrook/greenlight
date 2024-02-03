package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
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
type config struct {
	port int
	env  string
	db	 struct {
		dsn string
	}
}

// Define an app struct to hold dependencies.
// Dependencies:
//  1. config - the config struct
//  2. logger - System logger
type application struct {
	config config
	logger *log.Logger
}

// main function - The entry point for the app.
func main() {
	// Declare an instance of config struct
	var cfg config

	// Read command-line flags
	// Flags:
	//	1.	Server port (default: 4000)
	//	2.	Environment setting (default: development)
	//	3.	Database name (default: greenlight.db)
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "greenlight.db", "SQLite database name")
	flag.Parse()

	// Initialize a new logger that writes messages to
	// the standard out stream, prefixed with the
	// current date and time.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Call openDB() function to create connection pool,
	// passing in the config struct. If error returns,
	// log it and exit app immediately.
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// Defer db.Close() so connection pool closes before
	// the main() function exits
	defer db.Close()

	// Log message db is open
	logger.Printf("database connection pool established")

	// Declare an instance of the application struct.
	// Contains:
	//	1.	cfg struct
	//	2.	logger
	app := &application{
		config: cfg,
		logger: logger,
	}

	// Declare a new servermux.
	mux := http.NewServeMux()

	// Health Check route
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// Declare an HTTP server.
	// Parameters:
	//	1.	Addr: server port
	//	2.	Handler: use the httprouter in the routes file
	//	3.	IdleTimeout
	//	4.	ReadTimeout
	//	5.	WriteTimeout
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server.
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// openDB() function returns an sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	log.Printf("opening database: %v", cfg.db.dsn)
	db, err := sql.Open("sqlite3", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

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