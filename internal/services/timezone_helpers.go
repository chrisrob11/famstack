package services

import (
	"fmt"
	"time"
)

// ConvertToUTC converts a time from a specific timezone to UTC
func ConvertToUTC(t time.Time, timezone string) (time.Time, error) {
	if timezone == "" || timezone == "UTC" {
		return t.UTC(), nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	// If the time is naive (no timezone info), treat it as being in the specified timezone
	if t.Location() == time.UTC || t.Location().String() == "UTC" {
		// Create a new time in the specified timezone with the same date/time components
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc).UTC(), nil
	}

	// If it already has timezone info, convert to UTC
	return t.UTC(), nil
}

// ConvertFromUTC converts a UTC time to a specific timezone for display
func ConvertFromUTC(utcTime time.Time, timezone string) (time.Time, error) {
	if timezone == "" || timezone == "UTC" {
		return utcTime.UTC(), nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	return utcTime.In(loc), nil
}

// ConvertOptionalToUTC handles *time.Time fields for conversion to UTC
func ConvertOptionalToUTC(t *time.Time, timezone string) (*time.Time, error) {
	if t == nil {
		return nil, nil
	}

	converted, err := ConvertToUTC(*t, timezone)
	if err != nil {
		return nil, err
	}

	return &converted, nil
}

// ConvertOptionalFromUTC handles *time.Time fields for conversion from UTC
func ConvertOptionalFromUTC(utcTime *time.Time, timezone string) (*time.Time, error) {
	if utcTime == nil {
		return nil, nil
	}

	converted, err := ConvertFromUTC(*utcTime, timezone)
	if err != nil {
		return nil, err
	}

	return &converted, nil
}
