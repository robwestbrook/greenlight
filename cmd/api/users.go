package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/robwestbrook/greenlight/internal/data"
	"github.com/robwestbrook/greenlight/internal/validator"
)

// registerUserHandler registers a new user.
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Create an anonymous struct to hold the expected
	// data from the request body.
	var input struct {
		Name			string	`json:"name"`
		Email			string	`json:"email"`
		Password	string	`json:"password"`
	}

	// Parse the request body into the anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the data from the request body into a new
	// User struct. Set the activated field to false.
	user := &data.User{
		Name:				input.Name,
		Email: 			input.Email,
		Activated: 	false,
	}

	// Use the Password.Set() method to generate and
	// store the hashed and plaintext passwords.
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Create a new validator instance.
	v := validator.New()

	// Validate the user struct and return the error
	// messages to the client if any checks fail.
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the user data into the database.
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
			// If an ErrDuplicateEmail error is recieved,
			// use the v.AddError() method to manually add a
			// message to the validator instance, then call
			// our failedValidationResponse() helper.
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError(
				"email",
				"a user with this email address already exists",
			)
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write a JSON response containing the user data
	// along with a 201 Created status code.
	err = app.writeJSON(
		w,
		http.StatusCreated,
		envelope{"user": user},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// Create a background Goroutine to run mailer in 
	// a separate routine.
	go func() {

		// Run a deferred function which uses recover()
		// to catch any panic, and log an error message
		// instead of terminating the application.
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		// Call the Send() method on the Mailer, passing in
		// the user's email address, name of the template file,
		// and the User struct containing the new use's data.
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	}()
	
	// Send client a 202 Accepted status code. This status
	// code indicates the request has been accepted for
	// processing, but the processing has not been
	// completed.
	err = app.writeJSON(
		w,
		http.StatusAccepted,
		envelope{"user": user},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}