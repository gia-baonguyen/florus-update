package utils

import (
	"time"
)

// ParseDate parses a date string in various formats
func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, &time.ParseError{
		Layout:     "2006-01-02",
		Value:      dateStr,
		LayoutElem: "",
		ValueElem:  "",
		Message:    "unable to parse date",
	}
}
