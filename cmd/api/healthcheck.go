package main

import (
	"net/http"
)

// Declare a handler that writes a response with info
// about the app status, operating environment, and
// version.
// A METHOD on the APPLICATION struct.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create an envelope map, using the evelope type
	// in /cmd/api/helpers.go, that holds the information 
	// to send in the response
	env := envelope{
		"status":				"available",
		"system_info": map[string]string{
			"environment":	app.config.env,
			"version":			version,
		},
	}
	
	// Write JSON and headers using the writeJSON()
	// function in /cmd/api/helpers.go
	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(
			w,
			"The server encounted a problem and could not process your request",
			http.StatusInternalServerError,
		)
	}
}