package data

import (
	"database/sql"
	"errors"
)

// ErrRecordNotFound is returned when an event is
// not found in the database.
// ErrEditConflict is returned when a conflict race
// condition happens in the database.
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)

// Models is a struct which wraps all database models.
type Models struct {
	Events			EventModel
	Permissions	PermissionModel
	Tokens			TokenModel
	Users				UserModel
}

// NewModels returns a Models struct containing the
// initialized database models.
func NewModels(db *sql.DB) Models {
	return Models{
		Events: EventModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Tokens: TokenModel{DB: db},
		Users: UserModel{DB: db},
	}
}