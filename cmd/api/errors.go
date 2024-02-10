package main

import (
	"fmt"
	"net/http"
)

// logError method
// A METHOD on the APPLICATION struct.
func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

// errorResponse method
// Sends a JSON-formatted error message to a client with
// a given status code.
// A METHOD on the APPLICATION struct.
func (app *application) errorResponse(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	message interface{},
) {
	// Create an error envelope
	env := envelope{"error": message}

	// Write the response using the writeJSON() method.
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// serverErrorResponse method
// When the app encounters an unexpected problem at
// runtime, this method logs a detailed error message,
// then uses the errorResponse() method to send a
// 500 Internal Server Error status code and a JSON
// response to the client.
// A METHOD on the APPLICATION struct.
func (app *application) serverErrorResponse(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	app.logError(r, err)
	message := `
		the server encountered a problem and could not process your request
	`
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// notFoundResponse method
// Used to send a 404 Not found status code and JSON
// response to client
// A METHOD on the APPLICATION struct.
func (app *application) notFoundResponse(
	w http.ResponseWriter,
	r *http.Request,
) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// methodNotAllowedResponse method
// A METHOD on the APPLICATION struct.
func (app *application) methodNotAllowedResponse(
	w http.ResponseWriter,
	r *http.Request,
) {
	message := fmt.Sprintf(`
		the %s method is not supported for this resource
	`,
		r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// failedValidationResponse method
// Writes a 422 Unprocessable Entity and the contents
// of the error map.
func (app *application) failedValidationResponse(
	w http.ResponseWriter,
	r *http.Request,
	errors map[string]string,
) {
	app.errorResponse(
		w, 
		r, 
		http.StatusUnprocessableEntity, 
		errors)
}

// editConflictResponse method.
// Writes a 409 Conflict and plain English message.
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := `
		unable to update the record due to and edit
		conflict, please try again
	`
	app.errorResponse(w, r, http.StatusConflict, message)
}