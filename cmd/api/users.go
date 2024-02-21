package main

import (
	"errors"
	"net/http"
	"time"

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

	// Add the "events:read" permission for new user
	err = app.models.Permissions.AddForUser(
		user.ID,
		"events:read",
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// After the user record has been created in the
	// database, generate a new activation token for 
	// the user.
	token, err := app.models.Tokens.New(
		user.ID,
		3*24*time.Hour,
		data.ScopeActivation,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Use the background helper, found in helpers.go,
	// to execute an anonymous function that sends
	// the welcome email. Create a map as a holding
	// structure for the data passed to the email
	// template.
	app.background(func() {

		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID": user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	
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

// activateUserHandler is a method that received the
// client's activation token and activates the user
// in the database.
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Create an input struct to hold the plaintext
	// token.
	var input struct {
		TokenPlaintext	string	`json:"token"`
	}
	
	// Parse the plaintext activation token from the
	// request body.
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the plaintext token provided by client.
	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve the details of the user associated with
	// the token string using the GetForToken() method.
	// If no matching record is found, let the client
	// know the token provided is not valid.
	user, err := app.models.Users.GetForToken(
		data.ScopeActivation,
		input.TokenPlaintext,
	)
	
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Update the user's activation status and 
	// updated date and time.
	user.Activated = true
	user.UpdatedAt = time.Now()

	// Save the updated user record in the database,
	// checking for any edit conflicts.
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// If all is successful, delete all activation tokens
	// for the user.
	err = app.models.Tokens.DeleteAllForUser(
		data.ScopeActivation,
		user.ID,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send the updated user details to the client in
	// a JSON response.
	err = app.writeJSON(
		w,
		http.StatusOK,
		envelope{"user": user},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}