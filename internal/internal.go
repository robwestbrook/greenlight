package internal

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"
)

// dbTimeFormat defines the format used to convert
// date and time to a SQLite-friendly datetime.
const dbTimeFormat = "2006-01-02 15:04:05"

// StringToTime function takes in a time string 
// from SQLite. It returns a GO time.Time format.
// A METHOD on the APPLICATION struct.
func StringToTime(stringToConvert string) time.Time {
	// Only convert if the stringToConvert is not empty
	if stringToConvert != "" {
		res, _ := time.Parse(dbTimeFormat, stringToConvert)
		return res
	}
	return time.Time{}
}

// TimeToString function takes in the Go time.Time format
// and returns a time string for SQLite.
// A METHOD on the APPLICATION struct.
func TimeToString(timeToCovert time.Time) string {
	// Only convert if timeToConvert is not zero
	if !timeToCovert.IsZero() {
		t := time.Time(timeToCovert)
		return t.Format(dbTimeFormat)
	}
	return ""
}

// CurrentDate function generates a GO time.Time
// for the current date and time.
func CurrentDate() time.Time {
	return time.Now()
}

// StringToSlice converts a comma-delimited string 
// into a Go slice
func StringToSlice(str *[]string) *[]string {
	// Create a variable for the string slice passed in
	s := *str

	// Create a slice to hold the returned results
	var r []string

	// Loop over the string slice and append to returned
	// slice
	for _, x := range s {
		r = append(r, x)
	}

	// Return a pointer to the result slice
	return &r
}

// SliceToString converts a Go slice into a
// comma-delimited string
func SliceToString(s []string) string {
	if s != nil {
		return strings.Join(s, ",")
	}
	return ""
}

// GenerateRandomString generate a random string of 
// a supplied length.
func GenerateRandomString(length int) (string, error) {
	// Initialize a byte slice of given length.
	b := make([]byte, length)

	// Read the random bytes into b
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}