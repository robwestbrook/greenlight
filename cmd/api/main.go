package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Declare a string containing the app version.
const version = "1.0.0"

// Define a config struct to hold all configuration
// settings fot the app. These settings will be 
// read from command-line flags when the app starts.
// Configuration Settings:
//	1.	port - network port for server
//	2.	env - current operating environment
type config struct {
	port		int
	env			string
}

// Define an app struct to hold dependencies.
// Dependencies:
//	1.	config - the config struct
//	2.	logger - System logger
type application struct {
	config				config
	logger				*log.Logger
}

// main function - The entry point for the app.
func main() {
	// Declare an instance of config struct
	var cfg config

	// Read the port value and env command-line flags
	// into config struct. Default to port 4000 and
	// environment "development" when no flags are
	// provided.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialize a new logger that writes messages to
	// the standard out stream, prefixed with the
	// current date and time.
	logger := log.New(os.Stdout, "", log.Ldate | log.Ltime)

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
		Addr:					fmt.Sprintf(":%d", cfg.port),
		Handler: 			app.routes(),
		IdleTimeout: 	time.Minute,
		ReadTimeout: 	10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server.
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}