package data

import (
	"database/sql"
	"errors"
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

	// Use QueryRow() method to execute the SQL query
	// passing in the args slice. Scan in the returning
	// values to the event struct.
	return e.DB.QueryRow(query, args...).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt, &event.Version)
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

	// Execute the query with the QueryRow() method, 
	// passing the ID. Scan the response data into the
	// fields of the Event struct and tag variable.
	err := e.DB.QueryRow(query, id).Scan(
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
		WHERE id = ?
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
	}

	// Use QueryRow() method to execute query. Pass the 
	// args slice as a paramter and scan the new version
	// into the event struct.
	return e.DB.QueryRow(query, args...).Scan(&event.Version)
}

// Delete deletes a specific record by ID from 
// the events table.
func (e EventModel) Delete(id int64) error {
	return nil
}