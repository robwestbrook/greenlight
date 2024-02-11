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

// updateEventHandler updates a record in database
// A METHOD on the APPLICATION struct.
func (app *application) updateEventHandler(w http.ResponseWriter, r *http.Request) {
	// Extract event ID from the URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Get the existing event record from the database.
	// Send a 404 Not Found response if matching record
	// is not found.
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

	// Declare an input struct to hold data from client
	// Pointers have as their zero value: "nil". If an 
	// input field doesn't have any value, we can check
	// for zero value. The empty fields will be "nil". 
	var input struct {
		Title					*string		`json:"title"`
		Description		*string		`json:"description"`
		Tags 					[]string	`json:"tags"`
		AllDay				*bool			`json:"all_day"`
		Start					*string		`json:"start"`
		End						*string		`json:"end"`
	}

	// Read the JSON request body data into input struct.
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy values from request body to corresponding
	// fields of the event record.
	// If input values are nil, no corresponding
	// key/value pair was provided in the JSON request
	// body. Therefore no changes are made. Since the
	// input fields are pointers, they must be
	// dereferenced, using the * operator.
	if input.Title != nil {
		event.Title = *input.Title
	}
	if input.Description != nil {
		event.Description = *input.Description
	}
	// Tags are slices, which return nil if empty.
	// No need to dereference.
	if input.Tags != nil {
		event.Tags = input.Tags
	}
	if input.AllDay != nil {
		event.AllDay = *input.AllDay
	}
	if input.Start != nil {
		event.Start = internal.StringToTime(*input.Start)
	}
	if input.End != nil {
		event.End = internal.StringToTime(*input.End)
	}

	// Validate the updated event record. Send the client
	// a 422 Unprocessible Entity response if fails.
	v := validator.New()

	if data.ValidateEvent(v, event); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated event record to Update() method.
	// Check for edit conflict and server error
	err = app.models.Events.Update(event)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write the updated event record in a JSON response.
	err = app.writeJSON(
		w,
		http.StatusOK,
		envelope{"event": event},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteEventHandler deletes a record in database
// A METHOD on the APPLICATION struct.
func (app *application) deleteEventHandler(w http.ResponseWriter, r *http.Request) {
	// Get event ID from URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete event from database. Send a 404 Not Found
	// response to client if record not found.
	err = app.models.Events.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 200 OK status with success message
	err = app.writeJSON(
		w,
		http.StatusOK,
		envelope{"message": "event successfully deleted"},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// listEventsHandler returns multiple events to client.
// A METHOD on the APPLICATION struct.
func (app *application) listEventsHandler(w http.ResponseWriter, r *http.Request) {
	// Define an input struct to hold expected values
	// from the request query string.
	var input struct {
		Title				string
		Description	string
		Tags				[]string
		data.Filters
	}

	// Initialize a new Validator instance
	v := validator.New()

	// Call r.URL.Query() method to get the URL.Values
	// map containing the query string data.
	qs := r.URL.Query()

	// Use helpers to extract title and tags query string
	// values, falling back to defaults. Defaults:
	//	1.	title: ""
	//	2.	description: ""
	//	3.	tags: empty slice
	input.Title = app.readString(qs, "title", "")
	input.Description = app.readString(qs, "description", "")
	input.Tags = app.readCSV(qs, "tags", []string{})

	// Use helpers to extract page and page_size query
	// string values as integers. Read these values into
	// the embedded Filters struct. Defaults:
	//	1.	page: 1
	//	2.	page_size: 20
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// Use helpers to extract the sort query string value.
	// Read the value into the embedded Filters struct.
	// Default:
	//	1.	id
	input.Filters.Sort = app.readString(qs, "sort", "id")

	// Add supported values for sort to sort safelist
	input.Filters.SortSafelist = []string{
		"id",
		"title",
		"all_day",
		"start",
		"end",
		"-id",
		"-title",
		"-all_day",
		"-start",
		"-end",
	}
	
	// Execute the validation checks on the Filters
	// struct, sending a response containing errors.
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the GetAll() method to get events,
	// passing in filter parameters.
	events, metadata, err := app.models.Events.GetAll(
		input.Title,
		input.Description,
		input.Tags,
		input.Filters,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the event data.
	err = app.writeJSON(
		w,
		http.StatusOK,
		envelope{"events": events, "metadata": metadata},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}