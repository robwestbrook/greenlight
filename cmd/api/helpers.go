package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

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