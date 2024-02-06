package internal

import (
	"strings"
	"time"
)

// dbTimeFormat defines the format used to convert
// date and time to a SQLite-friendly datetime.
const dbTimeFormat = "2006-01-02 15:04:05"

// stringToTime function takes in a time string 
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

// timeToString function takes in the Go time.Time format
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

// current function generates a GO time.Time
// for the current date and time.
func CurrentDate() time.Time {
	return time.Now()
}

// stringToSlice converts a comma-delimited string 
// into a Go slice
func StringToSlice(s string) []string {
	if s != "" {
		return strings.Split(s, ",")
	}
	return nil
}

// sliceToString converts a Go slice into a
// comma-delimited string
func SliceToString(s []string) string {
	if s != nil {
		return strings.Join(s, ",")
	}
	return ""
}