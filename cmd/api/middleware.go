package main

import (
	"fmt"
	"net/http"
)

// recoverPanic sends an error when a panic occurs.
func (app *application) recoverPanic(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be
		// run in the event of a panic as Go unwinds the
		// stack.)
		defer func() {
			// Use the buildin recover function to check for
			// a panic.
			if err := recover(); err != nil {
				// If there is a panic, set a "Connection: close"
				// header on response. This acts as a trigger
				// to make Go's HTTP server automatically close
				// the current connection after a response is sent.
				w.Header().Set("Connection", "close")

				// The value returned by the recover() has the type
				// interface{}. Use fmt.Error() to normalize it into
				// an error and call the serverErrorResponse()
				// helper. This will log the error using the custom
				// Logger type at the ERROR level and send client a
				// 500 Internal Server Error response.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}