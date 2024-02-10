package data

import "github.com/robwestbrook/greenlight/internal/validator"

// Filters type
type Filters struct {
	Page					int
	PageSize			int
	Sort 					string
	SortSafelist	[]string
}

// ValidateFilters function performs sanity checks on
// the query string values.
func ValidateFilters(v *validator.Validator, f Filters) {
	// Check page,page_size, and sort contain sensible values
	// v.Check paramters:
	//	1.	ok: checks that statement is true or false
	//	2.	key: the parameter being validated
	//	3.	message: the message used when check fails
	v.Check(
		f.Page > 0, 
		"page", 
		"must be greater than zero",
	)
	v.Check(
		f.Page <= 10_000_000, 
		"page", 
		"must be a maximum of 10,000,000",
	)
	v.Check(f.PageSize > 0,
	"page_size",
	"must be greater than zero",
	)
	v.Check(
		f.PageSize <= 100,
		"page_size",
		"must be a maximum of 100",
	)
	v.Check(
		validator.In(f.Sort, f.SortSafelist),
		"sort",
		"invalid sort value",
	)
}