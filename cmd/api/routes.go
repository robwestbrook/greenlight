package main

/*
	Routing rules for the application.
*/

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// A METHOD on the APPLICATION struct.
func (app *application) routes() *httprouter.Router {
	// Initialize a new httprouter router instance
	router := httprouter.New()

	// Register the functions for the api endpoints.
	// Function attributes:
	//	1.	Request Method
	//	2.	URL path
	//	3.	Handler function

	// GET health check route
	router.HandlerFunc(
		http.MethodGet,
		"/v1/healthcheck",
		app.healthcheckHandler,
	)
	// POST create Event route
	router.HandlerFunc(
		http.MethodPost,
		"/v1/events",
		app.createEventHandler,
	)

	// GET get Event by ID route
	router.HandlerFunc(
		http.MethodGet,
		"/v1/events/:id",
		app.showEventHandler,
	)

	// Return the router instance
	return router
}