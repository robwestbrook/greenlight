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
	fmt.Fprintln(w, "create a new Event")
}

// showEventHandler
// A METHOD on the APPLICATION struct.
func (app *application) showEventHandler(w http.ResponseWriter, r *http.Request) {
	// Use the readIDParam method in helpers.go
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Create a new instance of the Event struct, containing
	// the ID extracted from URL.
	event := data.Event{
		ID: id,
		Title: "Work at Home Depot",
		Description: "This is a normal work day at Home Depot",
		Tags: []string{"work", "home depot"},
		AllDay: false,
		Start: app.stringToTime("2024-01-29 05:00:00"),
		End: app.stringToTime("2024-01-29 14:00:00"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version: 1,
	}

	// Encode the event struct to JSON and send it as
	// the HTTP response. Use the envelope type in
	// cmd/api/helpers.go to create an envelope instance
	// of the event.
	err = app.writeJSON(w, http.StatusOK, envelope{"event": event}, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(
			w,
			"The server encountered a problem and could not process your request",
			http.StatusInternalServerError,
		)
	}
}