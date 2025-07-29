package model

import (
	"encoding/json"
	"strings"
	"time"
)

// FlexibleTime is a custom time.Time type that can parse multiple date formats
type FlexibleTime time.Time

// UnmarshalJSON implements the json.Unmarshaler interface.
// This method allows parsing of multiple date formats from JSON
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var timeStr string
	err := json.Unmarshal(data, &timeStr)
	if err == nil {
		// If it's an empty string, just use zero time
		if strings.TrimSpace(timeStr) == "" {
			*ft = FlexibleTime(time.Time{})
			return nil
		}

		// Try to parse the time in different formats
		formats := []string{
			time.RFC3339,          // 2006-01-02T15:04:05Z07:00
			"2006-01-02",          // Simple date format
			"2006-01-02 15:04:05", // Date time format
			"2006/01/02",          // Alternative date format
			"01/02/2006",          // MM/DD/YYYY format
			"January 2, 2006",     // Month name format
			"2 January 2006",      // Day month format
		}

		for _, format := range formats {
			if t, err := time.Parse(format, timeStr); err == nil {
				*ft = FlexibleTime(t)
				return nil
			}
		}

		return &time.ParseError{
			Layout:     "multiple formats",
			Value:      timeStr,
			LayoutElem: "",
			ValueElem:  "",
			Message:    "could not parse time string in any of the supported formats",
		}
	}

	// If it's not a string, try to unmarshal as a time directly
	var t time.Time
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	*ft = FlexibleTime(t)
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
// This method ensures FlexibleTime is always marshaled in RFC3339 format
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	t := time.Time(ft)
	return json.Marshal(t.Format(time.RFC3339))
}

// String returns the time as a string in RFC3339 format
func (ft FlexibleTime) String() string {
	return time.Time(ft).Format(time.RFC3339)
}

// Time returns the underlying time.Time value
func (ft FlexibleTime) Time() time.Time {
	return time.Time(ft)
}
