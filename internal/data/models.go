package data

import (
	"database/sql"
	"errors"
)

// ErrRecordNotFound is returned when an event is
// not found in the database.
var (
	ErrRecordNotFound = errors.New("record not found")
)

// Models is a struct which wraps all database models.
type Models struct {
	Events			EventModel
}

// NewModels returns a Models struct containing the
// initialized database models.
func NewModels(db *sql.DB) Models {
	return Models{
		Events: EventModel{DB: db},
	}
}