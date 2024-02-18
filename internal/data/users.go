package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/robwestbrook/greenlight/internal"
	"github.com/robwestbrook/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// Define a custom ErrDuplicateEmail error.
var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// User defines a struct to represent an individual
// user. The json "-" struct tag is used to prevent the
// Password and Version fields from appearing in any
// output when enoding to JSON.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"-"`
}

// UserModel creates a struct that wraps the 
// connection pool.
type UserModel struct {
	DB *sql.DB
}

// password defines a struct which contains the
// plaintext and hashed versions of the user's password.
// The plaintext field is a pointer to a string, to
// distinguish between a plaintext password not being
// present in the struct at all, versus a plaintext
// password which is the empty string "".
type password struct {
	plaintext *string
	hash      []byte
}

// Insert a new record in the database for the user.
// Use the RETURNING clause to read the ID, created_at,
// and version into the Yser struct after the insert.
func (m UserModel) Insert(user *User) error {
	// Build the SQL query
	query := `
		INSERT INTO users (name, email, password_hash, activated, created_at, updated_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at, version
	`;

	// Create an args variable to hold the client input.
	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		internal.CurrentDate(),
		internal.CurrentDate(),
		1,
	}

	// Create a context with a 3 second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute SQL query.
	// If the table already contains a record with this
	// email address, an attempt to insert will be a
	// violation of the UNIQUE "users_email_key"
	// contraint. Check for this error and return custom
	// ErrDuplicateEmail error.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
	)
	if err != nil {
		switch {
			case err.Error() == `UNIQUE constraint failed: users.email`:
				return ErrDuplicateEmail
			default:
				return err
		}
	}
	return nil
}

// GetByEmail retrieves the User details from the
// database based on the user's email address. Because
// of the UNIQUE constaint on the email, the SQL query
// will return only one record, or none at all, where
// a ErrRecordNotFound error is returned.
func (m UserModel) GetByEmail(email string) (*User, error) {
	// Create a SQL query
	query := `
		SELECT id, name, email, password_hash, activated, created_at, updated_at, version
		FROM users
		WHERE email = ?
	`

	// Create a user variable to receive the database
	// response.
	var user User

	// Create a context with a 3 second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute SQL query.
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
	)

	// Check for errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

// Update the details for a specific user. Check
// against the version field to prevent any race
// conditions during the request cycle. Also check
// for a violation of the "user_email_key" constraint
// when performing the update.
func (m UserModel) Update(user *User) error {
	// Create SQL query.
	query := `
		UPDATE users
		SET name = ?, 
		email = ?, 
		password_hash = ?, 
		activated = ?, 
		created_at = ?, 
		updated_at = ?, 
		version = version + 1
		WHERE id = ? and version = ?
		RETURNING version
	`

	// Create an args variable to hold user input
	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.CreatedAt,
		user.UpdatedAt,
		user.ID,
		user.Version,
	}

	// Create a context with a 3 second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute database query
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.Version,
	)

	// Check for errors
	if err != nil {
		switch {
		case err.Error() == `Error: UNIQUE constraint failed`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Set method calculates the bcrypt hash of a
// plaintext password, and stores both the hash and the
// plaintext versions in the text.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(plaintextPassword),
		12,
	)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Matches method checks whether the provided
// plaintext password matches the hashed password
// stored in the struct. Return true if a match and
// false otherwise.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(
		p.hash,
		[]byte(plaintextPassword),
	)
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// GetForToken retrieves a user token from the database.
func (m UserModel) GetForToken(
	tokenScope string,
	tokenPlaintext string,
) (*User, error) {

	// Calculate the SHA-256 hash of the plaintext
	// token provided by the client.
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	// Compose the SQL query.
	// Use INNER JOIN to join together information from
	// the "users" and "tokens" tables. Use the 
	// "ON users.id = tokens.user_id" clause indicating
	// to join records where the user "id" value equals
	// the token "user_id".
	// "INNER JOIN" creates an interim table containing
	// the joined data from both tables. The "WHERE"
	// clause is used to filter this interim table to
	// leave only rows where the token hash and token
	// scope match specific placeholder parameter values,
	// and the token expiry is after a specific time.
	// Because the token hash is also a primary key,
	// only one record will be left which contains the
	// details of the user associated with the token
	// hash.
	query := `
		SELECT * FROM users
		INNER JOIN tokens
		ON users.id = tokens.user_id
		WHERE tokens.hash = ?
		AND tokens.scope = ?
		AND tokens.expiry > ?
	`

	// Create a slice ontaining the query arguments.
	// Use the [:] operator to get a slice containing
	// the token hash. Pass the current time as the
	// value to check against the expiry.
	args := []interface{}{
		tokenHash[:],
		tokenScope,
		time.Now(),
	}

	// Create a variable of type User
	var user User
	var token Token

	// Create a context with a 3 second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query, scanning the return values
	// into the User struct. If no match found, return
	// an ErrRecordNotFound error.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
		&token.Hash,
		&token.userID,
		&token.Expiry,
		&token.Scope,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	//**************
	// Check that token and token hash match.
	// if !CheckTokenHash(tokenPlaintext, token.Hash) {
	// 	return nil, errors.New("token is invalid or expired")
	// }
	//**************
	return &user, nil
}

// CheckTokenForHash function
func CheckTokenHash(token string, hash []byte) bool {
	log.Printf("Token: %s - Hash: %s", token, hash)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	return err == nil
}

// ValidateEmail creates validators for user email.
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(
		email != "",
		"email",
		"must be provided",
	)
	v.Check(
		validator.Matches(email, validator.EmailRX),
		"email",
		"must be a valid email address",
	)
}

// ValidatePasswordPlaintext create validators for
// user plaintext password.
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(
		password != "",
		"password",
		"must be provided",
	)
	v.Check(
		len(password) >= 8,
		"password",
		"must be at least 8 bytes long",
	)
	v.Check(
		len(password) <= 72,
		"password",
		"must not be more than 72 bytes long",
	)
}

// ValidateUser validates the user.
func ValidateUser(v *validator.Validator, user *User) {
	v.Check(
		user.Name != "",
		"name",
		"must be provided",
	)
	v.Check(
		len(user.Name) <= 500,
		"name",
		"must not be more than 500 bytes long",
	)

	// Call ValidateEmail() helper.
	ValidateEmail(v, user.Email)

	// If plaintext password is not nil, call
	// ValidatePasswordPlaintext() helper.
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	// If the password hash is ever nil, it will be due
	// to a logic error in the codebase. It is a useful
	// sanity check to include, but it is not a problem 
	// with the data peovided by the client. Rather than
	// adding an error to the validation map, raise a
	// panic instead.
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}
