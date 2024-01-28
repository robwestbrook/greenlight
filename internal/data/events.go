package data

import "time"

// Event struct
// Fields:
// 1.		ID: Unique ID for event
// 2.		Title: Event title
// 3.		Description: Event description
// 4.		Tags: Event tags
// 5.		AllDay: All day (true or false)
// 6.		Start: Start date and time
// 7.		End: End date and time
// 8.		CreatedAt: Timestamp when event was created
// 9.		UpdatedAt: Timestamp when event was updated
// 10.	Version: Version starts at 1 and incremented on each update
type Event struct {
	ID					int64				`json:"id"`
	Title				string			`json:"title"`
	Description	string			`json:"description,omitempty"`
	Tags				[]string		`json:"tags,omitempty"`
	AllDay			bool				`json:"all_day"`
	Start				time.Time		`json:"start"`
	End					time.Time	 	`json:"end"`
	CreatedAt		time.Time		`json:"created_at"`
	UpdatedAt		time.Time		`json:"updated_at"`
	Version			int32				`json:"version"`
}