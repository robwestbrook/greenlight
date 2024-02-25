package data

import (
	"context"
	"database/sql"
	"time"
)

// Permissions slice used to hold the permission
// code for a single user.
type Permissions []string

// PermissionModel defines the PermissionModel type
type PermissionModel struct {
	DB *sql.DB
}

// Include is a helper method to check if the
// Permissions slice contains a specific code.
func (p Permissions) Include(code string) bool {
	// Loop over permissions. If the slice contains
	// the parameter code, return true, otherwise false.
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

// GetAllForUser method returns all permission codes
// for a specific user in a Permissions slice.
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	// Compose query
	query := `
		SELECT permissions.code
		FROM permissions
		INNER JOIN users_permissions
		ON users_permissions.permission_id = permissions.id
		INNER JOIN users ON users_permissions.user_id = users.id
		WHERE users.id = ?
	`

	// Create a context with a 3 second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute query and recieve all rows
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create a permissions variable
	var permissions Permissions

	// Loop over all returned rows
	for rows.Next() {
		// create a permission variable
		var permission string

		// Scan row into permission variable
		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		// Append permission to the permissions variable
		permissions = append(permissions, permission)
	}
	// Check for errors
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// AddForUser adds provided codes for a specific user.
// TODO: process more than one code at a time.
func (m PermissionModel) AddForUser(
	userID int64,
	code string,
) error {

	// Build SQL query to get permissions ID from code
	query := `
		SELECT * FROM permissions
		WHERE code = ?
	`

	// Create a context with a 3 second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute query
	row := m.DB.QueryRowContext(ctx, query, code)

	// Initialize variables to hold query results
	var codeID int
	var codeString string

	// Scan the row results into the variables
	err := row.Scan(&codeID, &codeString)
	if err != nil {
		return err
	}

	// Build SQL query to insert userID and codeID
	// into users_permissions
	query = `
		INSERT INTO users_permissions (user_id, permission_id)
		VALUES (?, ?)
	`

	// Create a context with a 3 second timeout.
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute query
	_, err = m.DB.ExecContext(ctx, query, userID, codeID)
	return err
}
