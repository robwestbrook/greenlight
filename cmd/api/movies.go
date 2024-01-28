package main

import (
	"fmt"
	"net/http"
)

/*
	Handler Functions for Movies
*/

// createMovieHandler
// A METHOD on the APPLICATION struct.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

// showMovieHandler
// A METHOD on the APPLICATION struct.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Use the readIDParam method in helpers.go
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Show ID
	fmt.Fprintf(w, "show details of movie %d\n", id)
}