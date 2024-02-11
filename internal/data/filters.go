package data

import (
	"strings"

	"github.com/robwestbrook/greenlight/internal/validator"
)

// Filters type
type Filters struct {
	Page					int
	PageSize			int
	Sort 					string
	SortSafelist	[]string
}

// sortColumn function verifies the client-supplied
// Sort field matches with the safe list. If it does,
// extract the column name from the sort field.
func (f Filters) sortColumn() string {
	// Loop over the safelist
	for _, safeValue := range f.SortSafelist {
		// If the sort string equals a safelist entry
		// return the sort string, stipped of any "-".
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	// Failsafe to stop SQL injection
	panic("unsafe sort parameter: " + f.Sort)
}

// sortDirection function ("ASC" or "DESC") depending
// on prefix character of the Sort field.
func (f Filters) sortDirection() string {
	// If sort string has a "-" it is descending.
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	// If sort string has no "-", it is ascending.
	return "ASC"
}

// limit returns the page size
func (f Filters) limit() int {
	return f.PageSize
}

// offSet calculates the offset for pagination.
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
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