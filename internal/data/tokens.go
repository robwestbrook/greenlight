package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/robwestbrook/greenlight/internal/validator"
)

// Define constants for the token scope.
//	1.	Activation
//	2.	Authentication
const (
	ScopeActivation = "activation"
	ScopeAuthentication = "authenticaion"
)

// Token defines a struct to hold data for an individual
// token. This includes the plaintext and hashed
// versions of the token, associated userID, expiry
// time, and scope.
type Token struct {
	Plaintext 	string			`json:"token"`
	Hash 				[]byte			`json:"-"`
	userID			int64				`json:"-"`
	Expiry 			time.Time		`json:"expiry"`
	Scope 			string			`json:"-"`
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

	//******************
	// Use the GenerateRandomString() function from the 
	// internal package to return a random string.
	// HERE
	// randomString, err := internal.GenerateRandomString(24) 
	// if err != nil {
	// 	return nil, err
	// }
	//*****************

	// Initialize a zero-valued byte with a length of
	// 16 bytes.
	randomBytes := make([]byte, 16)

	//******************
	// Encode the byte slice to a base-32-encoded string
	// and assign it to the token Plainfield field. This
	// will be the token string sent to the user in the
	// welcome email.
	// token.Plaintext = randomString

	// hash, err :=bcrypt.GenerateFromPassword(
	// 	[]byte(token.Plaintext),
	// 	12,
	// )
	// if err != nil {
	// 	return nil, err
	// }

	// token.Hash = hash
	//*******************

	// Use the Read() function from the crypto/rand
	// package to fill the byte slice with random bytes
	// from the operating system's CSPRNG.
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Encode the byte slice to a base-32 encoded string
	// and assign it to the token Plaintext field. This
	// will be the token string sent to the user in the
	// welcome mail.
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Generate a SHA-256 hash of the plain text token
	// string. This will be the value stored in the hash
	// field of the database table.
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

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
	v.Check(
		len(tokenPlaintext) == 26,
		"token",
		"must be 26 bytes long",
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