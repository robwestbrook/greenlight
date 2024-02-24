package main

import (
	"errors"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/robwestbrook/greenlight/internal/data"
	"github.com/robwestbrook/greenlight/internal/validator"
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

// authenticate a user when a request is made.
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the
		// response. This indicates to any caches that the
		// response may vary based on the value of the
		// Authorization header in the request.
		w.Header().Add("Vary", "Authorization")

		// Retrieve the value of the Authorization header
		// from the request. If no header is found, an empty
		// string "" is returned.
		authorizationHeader := r.Header.Get("Authorization")

		// If no authorization header is found, use the
		// contextSetUser helper to add the Anonymous user
		// to the request context. Then call the next
		// handler in the chain.
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// If there is an authorization header, it is
		// expected to be in the format "Bearer <token>".
		// Split this format into its parts. If header is not
		// in expected format, return a 401 Unauthorized
		// response.
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Extract the actual authentication token.
		token := headerParts[1]

		// Validate the token to make use it is in a 
		//vsensible format.
		v := validator.New()

		// If the token is not valid, use the 
		// invalidAuthenticationTokenResponse() helper
		// to send a response.
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Retrieve the User details associated with the
		// authentication token.
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Call the contextSetUser() helper to add the user
		// information to the request context.
		r = app.contextSetUser(r, user)

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// requireActivatedUser checks if a user is
// authenticated and activated. It calls the 
// requireAuthenticatedUser before being executed
// itself.
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	// Assign the http.HandlerFunc to the variable fn
	// rather than return it.
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the contextGetUser() helper to retrieve the
		// user info from the request context.
		user := app.contextGetUser(r)

		// If the user is anonymous, call the
		// authenticationRequiredResponse() to inform the
		// client to authenticate before trying again.
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	// Wrap fn with the requiredAuthenticatedUser()
	// middleware before returning it.
	return app.requireAuthenticatedUser(fn)
}

// requireAuthenticatedUser middleware checks that a
// user is not anonymous.
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user info
		user := app.contextGetUser(r)

		// If the user is anonymous, send an authentication
		// is required response.
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		// If user is not anonymous, continue the 
		// middleware chain.
		next.ServeHTTP(w, r)
	})
}

// requirePermission middleware wraps the 
// requireActivatedUser middleware, which in turn,
// wraps the requireAuthenticated middleware. There
// will be three checks when this middleware runs.
// This middleware checks if an authenticated,
// activated user, has a specific permission.
// The first parameter is the permission code
// required for the user to have.
func (app *application) requirePermission(
	code string, 
	next http.HandlerFunc,
	) http.HandlerFunc {
		// Create a function to make this check
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Retrieve the user from the request context
			user := app.contextGetUser(r)

			// Get the slice of permissions for the user
			permissions, err := app.models.Permissions.GetAllForUser(user.ID)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			// Check of the slice includes the required 
			// permission. If not, then return a 403 Forbidden
			// status code.
			if !permissions.Include(code) {
				app.notPermittedResponse(w, r)
				return
			}

			// If the user has permission, call the next handler.
			next.ServeHTTP(w, r)
		}

		// Wrap this with the requireActivatedUser()
		// middleware before returning.
		return app.requireActivatedUser(fn)
}

// enableCORS method
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Origin" header
		w.Header().Add("Vary", "Origin")

		// Add the "Vary: Access-Control-Request-Method"
		// header.
		w.Header().Add("Vary", "Access-Control-Request-Method")

		// Get the value of the request's Origin header
		origin := r.Header.Get("Origin")

		// If there is an Origin request header present and
		// at least one trusted origin is configured.
		if origin != "" && len(app.config.cors.trustedOrigins) != 0 {
			// Loop through the list of trusted origins,
			// checking to see if the request origin exactly
			// matches on of them.
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					// If there is a match, set
					// "Access-Control-Allow-Origin" response
					// header with the request origin as the value.
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// Check if the request has the HTTP method OPTIONS
					// and contains "Access-Control-Request-Method"
					// header. If it is, treat as a preflight request.
					if r.Method == http.MethodOptions && 
						r.Header.Get("Access-Control-Request-Method") != "" {
							// Set necessary preflight response headers.
							w.Header().Set(
								"Access-Control-Allow-Methods",
								"OPTIONS, PUT, PATCH, DELETE",
							)
							w.Header().Set(
								"Access-Control-Allow-Headers",
								"Authorization, Content-Type",
							)
							// Write the headers along with a 200 OK
							// status and return from middleware with
							// no further action.
							w.WriteHeader(http.StatusOK)
							return
					}
				}
			}
		}
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// metrics method
func (app *application) metrics(next http.Handler) http.Handler {
	// Initialize new expvar variables when the middleware
	// chain is first built.
	//	1.	Total Requests Recieved
	//	2.	Total Responses Sent
	//	3.	Total Processing Time in microseconds
	//	4.	Count of Responses for each HTTP Status Code
	totalRequestsRecieved := expvar.NewInt("total_requests_recieved")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_us")
	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_Status")

	// this code runs on every request
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the add method to increment the number of
		// requests recieved by 1.
		totalRequestsRecieved.Add(1)

		// Call the httpsnoop.CaptureMetrics function,
		// passing in the next handler in the chain along
		// with the expisting http.ResponseWriter and
		// http.Request. This returns a metric struct.
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// On the way back up the middleware chain, increment
		// the number of responses sent by one.
		totalResponsesSent.Add(1)

		// Get request processing time in microsends from
		// httpsnoop and increment the cumulative
		// processing time.
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())

		// Use the Add() method to increment the count for
		// the given status code by one. The expvar map
		// is string-keyed, so use the strcov.Iota()
		// function to convert status code from integer 
		// to string.
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}