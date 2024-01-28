package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// dbTimeFormat defines the format used to convert
// date and time to a SQLite-friendly datetime.
const dbTimeFormat = "2006-01-02 15:04:05"

// Define a JSON envelope type
type envelope map[string]interface{}

/*
	Helper functions for the application
*/

// stringToTime function takes in a string defining the
// time format and a time string from SQLite. It returns
// a GO time.Time format.
// A METHOD on the APPLICATION struct.
func (app *application) stringToTime(stringToConvert string) (time.Time) {
	res, _ := time.Parse(dbTimeFormat, stringToConvert)
	return res
}

// Retrieve the "id" parameter from the current
// request context, convert it to an integer, and
// return. If not successful, return 0 and an error.
// A METHOD on the APPLICATION struct.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// Get URL parameters from ParamsFromContext() function
	// to get a slice containing all parameter names
	// and values.
	params := httprouter.ParamsFromContext(r.Context())

	// Use ByName() method to get the value of "id"
	// parameter from the params slice. 
	//
	// The value is always a string. Convert it to a base
	// 10 integer (64 bits). If it can't be converted, or is 
	// less than 1, the ID is invalid.
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	// Return the ID and nil error
	return id, nil
}

// writeJSON function creates JSON responses.
// Parameters:
//	1.	http.ResponseWriter
//	2.	HTTP status code to send
//	3.	Data to encode into JSON
//	4.	Header map containing additional HTTP headers
// A METHOD on the APPLICATION struct.
func (app *application) writeJSON(
	w http.ResponseWriter, 
	status int, 
	data envelope, 
	headers http.Header) error {
	// Encode the data to JSON, returning error if any
	// json.MarshallIndent() method makes JSON more
	// readable. Using no line prefix ("") and tab
	// indent ("\t") for each element.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Append new line to JSON for easier viewing
	js = append(js, '\n')

	// Add any headers, looping over the header map and 
	// adding each header to the http.ResponseWriter map.
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add "Content-Type: application/json" header, write
	// status code, and JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	// Return a nil error
	return nil

}