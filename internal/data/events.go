package data

import (
	"database/sql"
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

	// s := Scan{}

	// Use QueryRow() method to execute the SQL query
	// passing in the args slice. Scan in the returning
	// values to the event struct.
	return e.DB.QueryRow(query, args...).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt, &event.Version)
}

// Get fetches a specific record by ID from events table.
func (e EventModel) Get(id int64) (*Event, error) {
	return nil, nil
}

// Update updates a specific record by ID in 
// the events table.
func (e EventModel) Update(event *Event) error {
	return nil
}

// Delete deletes a specific record by ID from 
// the events table.
func (e EventModel) Delete(id int64) error {
	return nil
}