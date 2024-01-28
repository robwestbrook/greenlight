package main

import (
	"fmt"
	"net/http"
)

// Declare a handler that writes a response with info
// about the app status, operating environment, and
// version.
// A METHOD on the APPLICATION struct.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a fixed-format JSON response from a string.
	js := `{
			"status": "available",
			"environment": %q,
			"version": %q
	}`
	js = fmt.Sprintf(js, app.config.env, version)

	// Set "Content-Type: application/json" header on
	// response.
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON as the HTTP response body.
	w.Write([]byte(js))
}