package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/robwestbrook/greenlight/internal"
	"github.com/robwestbrook/greenlight/internal/data"
	"github.com/robwestbrook/greenlight/internal/validator"
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

	// Copy values from input struct to a new Event struct
	event := &data.Event{
		Title: input.Title,
		Description: input.Description,
		Tags: input.Tags,
		AllDay: input.AllDay,
		Start: internal.StringToTime(input.Start),
		End: internal.StringToTime(input.End),
	}

	// Initialize a new Validator
	v := validator.New()

	// Call the ValidateEvent() function and return a
	// response contianing errors if any checks fail
	if data.ValidateEvent(v, event); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Events model Insert() method, passing a
	// pointer to the validated event struct. This
	// creates a record in the database and updates the
	// event struct with system-generated info.
	err = app.models.Events.Insert(event)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// With the HTTP response, include a Location header
	// so the client knows which URL to find the resource.
	// Create an empty http.Header map  and use the Set()
	// method to add a new Location header, using the ID
	// in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/events/%d", event.ID))

	// Write a JSON response with a 201 Created status code,
	// the event in the response body, and the location
	// header.
	err = app.writeJSON(
		w,
		http.StatusCreated,
		envelope{"event": event},
		headers,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
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

	// Call the Get() method to fetch data for a specific
	// event. Use Errors.Is() to check if a 
	// data.ErrRecordNotFound is returned. If so, send a
	// 404 Not Found response to client.
	event, err := app.models.Events.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
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
