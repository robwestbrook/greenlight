package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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
	// Define a client struct to hold the rate limiter
	// and last seen time for each client.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Create a mutex and map to hold the client struct.
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine that removes old
	// entries from the clients map once every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter
			// checks while cleanup is taking place.
			mu.Lock()

			// Loop through all clients. If they have not
			// made any requests within the last 3 minutes,
			// delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// Unlock mutex when cleanup is complete.
			mu.Unlock()
		}
	}()

	// The function returned is a closure, which
	// "closes over" the limiter variable.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only rate limit if it is enabled
		if app.config.limiter.enabled {
			// Extract the client's IP address from the request.
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			// Lock the mutex to prevent this code from being
			// executed concurrently.
			mu.Lock()

			// Check if IP address already exists in the client
			// map. If not, initialize a new rate limiter and
			// add the IP address and limiter to the map.
			// Limiter parameters:
			//	1.	no more than average of 2 requests per second.
			//	2.	maximum of 4 requests in a "burst".
			if _, found := clients[ip]; !found {

				// Create and add a new client struct to the map
				// if it does not already exist. Use the
				// requests per second and burst values from
				// the config struct.
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps),
						app.config.limiter.burst),
				}
			}
			// Update the last seen time for the client.
			clients[ip].lastSeen = time.Now()

			// Call the limiter.Allow() method to check if the request is
			// permitted. Whenever the limiter.Allow() method
			// is called, exactly one token will be consumed
			// from the bucket. If no tokens are left in the
			// bucket, call the rateLimitExceededResponse()
			// helper to return a 429 Too Many Requests response.
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// Unlock the mutex before calling the next handler
			// in the chain. DO NOT defer to unlock the mutex.
			// This would mean the mutex is not unlocked until
			// all handlers downstream of this middleware have
			// also returned.
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}
