package validator

import "regexp"

// Define a regular expression for checking email
// addresses.
var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zAZ0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	)

// Validator defines a struct that contains a map
// of validation errors.
type Validator struct {
	Errors				map[string]string
}

// New function is a helper that creates a new
// Validator instance with an empty errors map.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid function returns true if the errors map
// doesn't have any entries.
// A METHOD on Validator
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError function adds an error message to the Error
// map. (As long as no entry already exists for given key)
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check function adds an error message to the map only
// if a validation check is not "ok".
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Matches function returns true if a string value
// matches a specific regexp pattern.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique function returns true if all string values
// in a slice is unique.
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}