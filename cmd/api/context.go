package main

import (
	"context"
	"net/http"

	"github.com/robwestbrook/greenlight/internal/data"
)

// Define a custom contextKey type as a string.
type contextKey string

// Convert the string "user" to a contextKey type
// and assign it to the userContextKey constant. This
// constant is used as the key for getting and setting
// user info in the request content.
const userContextKey = contextKey("user")

// contextSetUser method returns a new copy of the
// request with the provided User struct added to the
// context. Use userContextKey as the key.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contextGetUser method retrieves the User struct from
// the request context. Only used when a User struct
// value is expected in the context. If it does not, it
// will firmly be an "unexpected" error. 
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}