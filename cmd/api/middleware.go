package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

// recoverPanic sends an error when a panic occurs.
// This is a MIDDLEWARE METHOD for application.
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

// rateLimit limits the rate of using the API
// This is a MIDDLEWARE METHOD for application.
func (app *application) rateLimit(next http.Handler) http.Handler {
	// Initialize a new rate limiter. Parameters:
	//	1.	Average of 2 requests per second.
	//	2.	Maximum of 4 requests in a single "burst".
	limiter := rate.NewLimiter(2, 4)

	// The function returned is a closure, which
	// "closes over" the limiter variable.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Call limiter.Allow() to check if the request is
		// permitted. Whenever the limiter.Allow() method
		// is called, exactly one token will be consumed
		// from the bucket. If no tokens are left in the
		// bucket, call the rateLimitExceededResponse()
		// helper to return a 429 Too Many Requests response.
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}