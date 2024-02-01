package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/robwestbrook/greenlight/internal/data"
)

/*
	Handler Functions for Events
*/

// createEventHandler
// A METHOD on the APPLICATION struct.
func (app *application) createEventHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous strut to hold the info
	// expected in the HTTP body. This struct is the
	// *target decode destination*.
	var input struct {
		Title       string   `json:"title"`
		Description string   `json:"description,omitempty"`
		Tags        []string `json:"tags,omitempty"`
		AllDay      bool     `json:"all_day"`
		Start       string   `json:"start"`
		End         string   `json:"end"`
	}

	// Use the readJSON() helper to decode request body
	// into the input struct. If an error is returned,
	// use the badRequestResponse() helper.
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Place the contents of the input struct into an
	// HTTP response.
	fmt.Fprintf(w, "%+v\n", input)
}

// showEventHandler
// A METHOD on the APPLICATION struct.
func (app *application) showEventHandler(w http.ResponseWriter, r *http.Request) {
	// Use the readIDParam method in helpers.go
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Create a new instance of the Event struct, containing
	// the ID extracted from URL.
	event := data.Event{
		ID:          id,
		Title:       "Work at Home Depot",
		Description: "This is a normal work day at Home Depot",
		Tags:        []string{"work", "home depot"},
		AllDay:      false,
		Start:       app.stringToTime("2024-01-29 05:00:00"),
		End:         app.stringToTime("2024-01-29 14:00:00"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
	}

	// Encode the event struct to JSON and send it as
	// the HTTP response. Use the envelope type in
	// cmd/api/helpers.go to create an envelope instance
	// of the event.
	err = app.writeJSON(w, http.StatusOK, envelope{"event": event}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
