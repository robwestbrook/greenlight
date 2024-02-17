package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/robwestbrook/greenlight/internal/validator"
)

// Define a JSON envelope type
type envelope map[string]interface{}

/*
	Helper functions for the application
*/

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
func (app *application) readJSON(
	w http.ResponseWriter, 
	r *http.Request, 
	dst interface{},
	) error {
		// Use http.MaxBytesReader() to limit the size of
		// the request body to 1MB.
		maxBytes := 1_048_576
		r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

		// Initialize the json.Decoder, and call the
		// DisallowedUnknownFields() method before decoding.
		// If the JSON from the client includes any fields
		// that can't be mapped to the target destination,
		// the decoder will return an error instead of
		// ignoring the field.
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		// Decode the request body to the destination.
		err := dec.Decode(dst)
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

			// If JSON contains a field that cannot be mapped
			// to the target destination, then Decode() will
			// return an error message in the format:
			// "json: unknown field <name>".
			case strings.HasPrefix(err.Error(), "json: unknown field "):
				fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
				return fmt.Errorf("body contains unknown key %s", fieldName)

			// If request body exceeds 1MB in size the decode
			// will fail with the error "http: request body
			// to large".
			case err.Error() == "http: request body too large":
				return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

			// A json.InvalidUnmarshalError error will be returned
			// if a non-nil pointer is passed to Decode(). Catch
			// and panic. Fail fast on errors that shouldn't occur
			// during normal operation that can't be handled
			// gracefully.
			case errors.As(err, &invalidUnmarshalError):
				panic(err)

			// For anything else, return error message as is.
			default:
				return err
			}
		}

		// Call Decode() again, using a pointer to an empty
		// anonymous struct as the destination. If the
		// request body only contains a single JSON value,
		// return an io.EOF error. If anything else is
		// returned, there is additional data in the request
		// body. Return a custom error message.
		err = dec.Decode(&struct{}{})
		if err != io.EOF {
			return errors.New("body must only contain a single JSON value")
		}

		return nil
}


// readString helper function returns a string value
// from the query string, or the provided default value
// if no matching key is found.
// A METHOD on the APPLICATION struct.
func (app *application) readString(
	qs url.Values,
	key string,
	defaultValue string,
) string {
	// Extract the value for a given key from the query
	// string. If no key exists, return the empty string ""
	s := qs.Get(key)

	// If no key exists, return the default value.
	if s == "" {
		return defaultValue
	}

	// If there is a key, return string
	return s
}

// readCSV helper function reads a string value from the
// query string and splits it into a slice on the comma
// character. If no matching key is found, returns
// the provided default value.
// A METHOD on the APPLICATION struct.
func (app *application) readCSV(
	qs url.Values,
	key string,
	defaultValue []string,
) [] string {
	// Extract the value for the key
	csv := qs.Get(key)

	// If no key exists, return the default value.
	if csv == "" {
		return defaultValue
	}

	// Parse the value into a []string slice and return
	return strings.Split(csv, ",")
}

// readInt helper function reads a string value from 
// the query string and converts it to an integer
// before returning. If no key is found, return
// default value. If value cannot be converted to an
// integer, record an error message to Validator
// instance.
// A METHOD on the APPLICATION struct.
func (app *application) readInt(
	qs url.Values,
	key string,
	defaultValue int,
	v *validator.Validator,
) int {
	// Extract the value of key
	s := qs.Get(key)

	// If no key exists, return the default value.
	if s == "" {
		return defaultValue
	}

	// Convert value to an integer. If this fails, add
	// an error message to validator instance and return
	// default value.
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	// Return converted integer value.
	return i
}

// background is a helper function that wraps
// panic recovery logic. The function accepts
// an arbitrary function as a parameter.
func (app *application) background(fn func()) {
	// Launch a background goroutine.
	go func() {
		// Recover any panic.
		defer func ()  {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		fn()
	}()
}