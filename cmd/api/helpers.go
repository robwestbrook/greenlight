package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
func (app *application) stringToTime(stringToConvert string) time.Time {
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
//  1. http.ResponseWriter
//  2. HTTP status code to send
//  3. Data to encode into JSON
//  4. Header map containing additional HTTP headers
//
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

// readJSON helper function will decode the JSON from
// the request body, then triage the errors and replace
// them with custom messages.
// A METHOD on the APPLICATION struct.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Decode request body into target destination
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {

		// Check for errors. If there is one during decoding,
		// start the triage.
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch {
		// Use the Errors.As() function to check if the
		// error has a type *json.SyntaxError. If it does,
		// return a plain-english error message	which
		// includes the location of the problem.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// Decode() may return an io.ErrUnexpectedEOF error
		// for syntax errors in the JSON. Check for this
		// using errors.Is() and return a generic error
		// message.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// Catch any *json.UnmarshalTypeError errors. These
		// occur when the JSON value is the wrong type for
		// the target destination. If the error relates to a
		// specific field, include in error message.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// An io.EOF error is returned by Decode() if the
		// request body is empty. Check for this with
		// errors.Is() an return plain-english error message.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// A json.InvalidUnmarshalError error will be returned
		// if a non-nil pointer is passed to Decode(). Catch
		// and panic.
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// For anything else, return error message as is.
		default:
			return err
		}
	}
	return nil
}
