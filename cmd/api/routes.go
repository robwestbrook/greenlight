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

	// Convert the notFoundResponse() helper to an
	// http.Handler using the http.HandleFunc adapter,
	// and then set it as the custom error handler for
	// 404 Not Found responses.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// Convert the methodNotAllowedResponse helper to an
	// http.Handler and set it as the custom error handler
	// for 405 Method Not Allowed responses.
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Register the functions for the api endpoints.
	// Function attributes:
	//	1.	Request Method
	//	2.	URL path
	//	3.	Handler function

	// GET health check route
	// Pattern					|		Handler						|		Action
	//----------------------------------------------------
	// /v1/healthcheck	|	healthcheckHandler	| show app
	//									|											| information
	router.HandlerFunc(
		http.MethodGet,
		"/v1/healthcheck",
		app.healthcheckHandler,
	)
	// GET list events route
	// Pattern					|		Handler						|		Action
	//----------------------------------------------------
	// /v1/events				|	listEventsHandler		| retrieve list
	//									|											| of events
	router.HandlerFunc(
		http.MethodGet,
		"/v1/events",
		app.listEventsHandler,
	)
	// POST create Event route
	// Pattern					|		Handler						|		Action
	//----------------------------------------------------
	// /v1/events				|	createEventHandler	| create new
	//									|											| event
	router.HandlerFunc(
		http.MethodPost,
		"/v1/events",
		app.createEventHandler,
	)

	// GET get Event by ID route
	// Pattern					|		Handler						|		Action
	//----------------------------------------------------
	// /v1/events	/:id	|	showEventHandler		| show event
	//									|											| details
	router.HandlerFunc(
		http.MethodGet,
		"/v1/events/:id",
		app.showEventHandler,
	)

	// PATCH update Event by ID route
	// Pattern					|		Handler						|		Action
	//----------------------------------------------------
	// /v1/events	/:id	|	updateEventHandler	| update event
	//									|											| details
	router.HandlerFunc(
		http.MethodPatch,
		"/v1/events/:id",
		app.updateEventHandler,
	)

	// DELETE delete Event by ID
	// Pattern					|		Handler						|		Action
	//----------------------------------------------------
	// /v1/events	/:id	|	deleteEventHandler	| delete event
	//									|											| by ID
	router.HandlerFunc(
		http.MethodDelete,
		"/v1/events/:id",
		app.deleteEventHandler,
	)

	// Return the router instance
	return router
}
