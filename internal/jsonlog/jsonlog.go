package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Level efines a level type to represent the security
// level for a log entry.
type Level int8

// Initialize constants for specific security levels.
// Use iota to assign successive integer values to
// the constants.
const (
	LevelInfo  Level = iota // value 0
	LevelError              // value 1
	LevelFatal              // value 2
	LevelOff                // value 3
)

// Return a human-friendly string for severity level.
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Logger defines a custom logger type. This type holds:
//  1. Output destination
//  2. Minimum severity level entries written for
//  3. Mutex for coordinating the writes. A mutex is
//     a mutual exclusion lock. This prevents the
//     logger from making multiple writes concurrently.
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// New returns a new Logger instance that writes log entries
// at or above a minimum severity level to a specific
// output destination.
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// HELPER METHODS

// PrintInfo logs application information.
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

// PrintError logs application errors.
func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

// PrintFatal logs fatal application errors and
// terminates the application.
func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	// Fatal level terminates the application.
	os.Exit(1)
}

// Print writes the log entry.
func (l *Logger) print(
	level Level,
	message string,
	properties map[string]string,
) (int, error) {
	// If the severity level is below the minimum
	// severity, return with no further action.
	if level < l.minLevel {
		return 0, nil
	}

	// Define an anonymous struct holding data fot
	// log entry.
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	// Include stack trace for entries at the ERROR
	// and FATAL levels.
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	// Declare a line variable for holding actual
	// log entry text.
	var line []byte

	// Marshall the anonymous struct to JSON and store
	// in the line variable. If creating JSON errors,
	// set contents of log entry to plain-English
	// error message.
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshall log message:" + err.Error())
	}

	// Lock the mutex so no two entries write to output
	// destination concurrently.
	l.mu.Lock()
	defer l.mu.Unlock()

	// Write the log entry followed by a newline.
	return l.out.Write(append(line, '\n'))
}

// Write method implemented to satisfy the io.Writer
// interface.
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
