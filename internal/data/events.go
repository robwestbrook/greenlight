package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/robwestbrook/greenlight/internal"
	"github.com/robwestbrook/greenlight/internal/validator"
)

// Event struct
// Fields:
// 1.		ID: Unique ID for event
// 2.		Title: Event title
// 3.		Description: Event description
// 4.		Tags: Event tags
// 5.		AllDay: All day (true or false)
// 6.		Start: Start date and time
// 7.		End: End date and time
// 8.		CreatedAt: Timestamp when event was created
// 9.		UpdatedAt: Timestamp when event was updated
// 10.	Version: Version starts at 1 and incremented on each update
type Event struct {
	ID					int64				`json:"id"`
	Title				string			`json:"title"`
	Description	string			`json:"description,omitempty"`
	Tags				[]string		`json:"tags,omitempty"`
	AllDay			bool				`json:"all_day"`
	Start				time.Time		`json:"start"`
	End					time.Time	 	`json:"end"`
	CreatedAt		time.Time		`json:"created_at"`
	UpdatedAt		time.Time		`json:"updated_at"`
	Version			int32				`json:"version"`
}

// EventModel struct wraps an sql.DB connection pool.
type EventModel struct {
	DB 	*sql.DB
}

// ValidateEvent runs the validator to validate
// events
func ValidateEvent(v *validator.Validator, event *Event) {
	v.Check(event.Title != "", "title", "must be provided")
	v.Check(len(event.Title) < 100, "title", "must not be more than 100 bytes long")
	v.Check(len(event.Description) <= 500, "description", "must not be more than 500 bytes long")
	v.Check(!event.Start.IsZero() || event.AllDay, "start", "if all day is false start must have a date")
}

// Insert a new record into the events table.
func (e EventModel) Insert(event *Event) error {
	// Define the SQL query for inserting a new record
	// in the events table, returning the system
	// generated data.
	query := `
		INSERT INTO events (title, description, tags, all_day, start, end, created_at, updated_at, version)
		VALUES (?, ? ,?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at, version;
	`

	// Create an arguments slice containing the values
	// for the placeholder parameters.
	args := []interface{}{
		event.Title,												// title - string
		event.Description,									// description - string
		internal.SliceToString(event.Tags),	// tags - string
		event.AllDay,												// all_day - boolean
		event.Start,												// start - convert from Go time to string
		event.End,													// end - convert from Go time to string
		time.Now(), 												// created_at - convert from Go time to string
		time.Now(),													// updated_at - convert from Go time to string
		1,																	// version - starts with 1
	}

	// Create a context with a 3 second timeout and defer.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryRowContext() method to execute the SQL query
	// passing in the context, query, and args slice. 
	// Scan in the returning values to the event struct.
	return e.DB.QueryRowContext(ctx, query, args...).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt, &event.Version)
}

// Get fetches a specific record by ID from events table.
func (e EventModel) Get(id int64) (*Event, error) {
	// Check that ID is not less than 1
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving event data
	query := `
		SELECT *
		FROM events
		WHERE id = ?
	`

	// Declare an Event struct to hold returned data
	var event Event

	// Declare a tag string variable to hold returned
	// tags value. The tags are stored in the SQLite
	// database as a comma-delimited string.
	var tags string

	// Use the context.WithTimeout() function to create
	// a context.Context which carries a 3 second
	// timeout deadline. Use the empty context.Background
	// as the "parent" context.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Use defer to make sure the context is cancelled
	// before the Get() method returns.
	defer cancel()

	// Execute the query with the QueryRowContext() method, 
	// passing the  context with deadline and ID. 
	// Scan the response data into the fields of the 
	// Event struct and tag variable.
	err := e.DB.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&tags,
		&event.AllDay,
		&event.Start,
		&event.End,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.Version,
	)

	// Convert tags to slice and add to event.Tags struct
	event.Tags = strings.Split(tags, ",")

	// If no matching event found, Scan() returns an
	// sql.ErrNoRows error. Check and return custom
	// ErrRecordNotFound error.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// If no errors, return pointer to Event struct.
	return &event, nil
}

// Update updates a specific record by ID in 
// the events table.
func (e EventModel) Update(event *Event) error {
	// Define the SQL query to update event
	query := `
		UPDATE events
		SET 
		title = ?, 
		description = ?, 
		tags = ?, 
		all_day = ?,
		start = ?,
		end = ?,
		updated_at = ?,
		version = version + 1
		WHERE id = ? AND version = ?
		RETURNING version
	`

	// Create a args slice containing the values for the
	// placeholder parameters.
	args := []interface{}{
		event.Title,
		event.Description,
		internal.SliceToString(event.Tags),
		event.AllDay,
		event.Start,
		event.End,
		internal.CurrentDate(),
		event.ID,
		event.Version,
	}

	// Create a context with a 3 second timeout and defer.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryRowContext() method to execute query. 
	// Pass the context, query, and args slice as paramters 
	// and scan the new version into the event struct. 
	// If no row is found, the  event has been deleted or 
	// the version has changed, indicating a race condition.
	err := e.DB.QueryRowContext(ctx, query, args...).Scan(&event.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete deletes a specific record by ID from 
// the events table.
func (e EventModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if event ID
	// is less than 1
	if id < 1 {
		return ErrRecordNotFound
	}

	// Build SQL query
	query := `
		DELETE FROM events
		WHERE id = ?
	`

	// Create a context with a 3 second timeout and defer.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query using the Exec() method, passing
	// in the context, query, and ID.
	result, err := e.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result
	// object to get number of rows affected by query.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows affected, the events table did not
	// contain a record with the ID. Return an
	// ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// GetAll() method returns a slice of events.
func (e EventModel) GetAll(
	title string,
	description string,
	tags	[]string,
	filters 	Filters,
) ([]*Event, error) {
	// Build the SQL query to get all event records
	query := fmt.Sprintf(`
		SELECT *
		FROM events
		WHERE (
			INSTR(LOWER(title), LOWER(?)) 
			OR ? = ''
		)
		AND INSTR(LOWER(description), LOWER(?))
		AND INSTR(tags, ?) 
		ORDER BY %s %s, id ASC
	`,
	filters.sortColumn(), 
	filters.sortDirection(),
	)

	// Create a context with 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryContext() method to execute the query.
	// An sql.Rows result set is returned containing
	// the result. QueryContext Paramters:
	//	1.	ctx: context
	//	2.	query: query string
	//	3.	title: title passed in to function (used twice)
	//	4.	title: title passed in to function (used twice)
	//	5.	description: description passed in to function
	//	6.	tags: convert tag slice passed in to string
	rows, err := e.DB.QueryContext(
		ctx, 
		query, 
		title, 
		title, 
		description,
		internal.SliceToString(tags),
	)
	if err != nil {
		return nil, err
	}

	// Defer a call to rows.Close()
	defer rows.Close()

	// Initialize an empty slice to hold event data
	events := []*Event{}

	// Use rows.Next to iterate over the rows in the
	// result set.
	for rows.Next() {
		// Initialize an empty Event struct to hold
		// each event
		var event Event

		// Initialize an empty Tag slice to hold 
		// event tags
		var tags string

		// Scan values into movie struct.
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&tags,
			&event.AllDay,
			&event.Start,
			&event.End,
			&event.CreatedAt,
			&event.UpdatedAt,
			&event.Version,
		)

		// Convert tags to slice and add to event.Tags 
		// struct
		event.Tags = strings.Split(tags, ",")

		if err != nil {
			return nil, err
		}

		// Add Event struct to the events slice
		events = append(events, &event)
	}

	// After rows.Next() loop is finished, call rows.Err()
	// to get any error encountered during loop.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// If everything goes OK, return slice of events.
	return events, nil
}