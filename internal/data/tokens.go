package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/robwestbrook/greenlight/internal"
	"github.com/robwestbrook/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// Define constants for the token scope.
const (
	ScopeActivation = "activation"
)

// Token defines a struct to hold data for an individual
// token. This includes the plaintext and hashed
// versions of the token, associated userID, expiry
// time, and scope.
type Token struct {
	Plaintext 	string
	Hash 				[]byte
	userID			int64
	Expiry 			time.Time
	Scope 			string
}

// TokenModel defines the TokenModel type.
type TokenModel struct {
	DB *sql.DB
}

// generateToken function generates a token.
func generateToken(
	userID int64,
	ttl time.Duration,
	scope string,
) (*Token, error) {
	//	Create a token containing the user ID,
	// expiry, and scope information.
	token := &Token{
		userID: userID,
		Expiry: time.Now().Add(ttl),
		Scope: scope,
	}

	// Use the GenerateRandomString() function from the 
	// internal package to return a random string.
	randomString, err := internal.GenerateRandomString(24) 
	if err != nil {
		return nil, err
	}

	// Encode the byte slice to a base-32-encoded string
	// and assign it to the token Plainfield field. This
	// will be the token string sent to the user in the
	// welcome email.
	token.Plaintext = randomString

	hash, err :=bcrypt.GenerateFromPassword(
		[]byte(token.Plaintext),
		12,
	)
	if err != nil {
		return nil, err
	}

	token.Hash = hash

	return token, nil
}

// ValidateTokenPlaintext checks that the plaintext
// token has been provided and is exactly 52 bytes long.
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(
		tokenPlaintext != "",
		"token",
		"must be provided",
	)
}

// New method is a shortcut which creates a new Token
// struct and inserts the data in the tokens table.
func (m TokenModel) New(
	userID int64,
	ttl time.Duration,
	scope string,
) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

// Insert method adds the data for the specific token
// to the tokens table.
func (m TokenModel) Insert(token *Token) error {
	// Create SQL query
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES (?, ?, ?, ?)
	`

	// Create an args variable to hold the values
	args := []interface{}{
		token.Hash,
		token.userID,
		token.Expiry,
		token.Scope,
	}

	// Create a context with 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute query
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// DeleteAllForUser deletes all tokens for a specific
// user and scope.
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	// Create SQL query
	query := `
		DELETE FROM tokens
		WHERE scope = ? AND user_id = ?
	`

	// Create a context with 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute query
	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}