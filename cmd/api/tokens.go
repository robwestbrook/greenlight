package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/robwestbrook/greenlight/internal/data"
	"github.com/robwestbrook/greenlight/internal/validator"
)

// createAuthenticationTokenHandler creates an
// authentication token. This handler exchanges
// the user's email address and password for an
// authentication token.
func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Create variables to hold client input.
	var input struct {
		Email			string	`json:"email"`
		Password	string	`json:"password"`
	}

	// Parse the email and password from the
	// request body.
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Create a new validator instance and validate the
	// client's email and password.
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Lookup the user record based on email address. If
	// no match found, call app.invalidCredentialsResponse
	// helper to send a 401 Unauthorized response to 
	// the client.
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Check if the provided password matches the actual
	// password for the user.
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// If the passwords don't match, call the 
	// app.invalidCredentialsResponse() helper and return.
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// If password is correct, generate a new token
	// with a 24 hour expiry time and scope "authentication".
	token, err := app.models.Tokens.New(
		user.ID,
		24*time.Hour,
		data.ScopeAuthentication,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Encode the token to JSON and send it in the
	// response along with a 201 Created status code.
	err = app.writeJSON(
		w,
		http.StatusCreated,
		envelope{"authentication_token": token},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}