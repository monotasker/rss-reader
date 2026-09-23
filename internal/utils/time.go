// Package utils provides utilities for rss-reader.
//
// Includes tools and helpers for:
//   - time parsing and formatting
package utils

import (
	"fmt"
	"time"
)

const timeFormat = time.RFC3339

// Formats previously written by binding time.Time directly to SQLite (space
// separator + fractional seconds) instead of FmtTime.
var parseFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02 15:04:05Z07:00",
}

// FmtTime formats a given time to RFC3339 format (e.g, 2026-10-28T17:58:45Z).
// This is the format used for inserting times into the db.
func FmtTime(t time.Time) string { return t.UTC().Format(timeFormat) }

// FmtTimePointer converts an optional time to an optional string
func FmtTimePointer(t *time.Time) *string {
	if t == nil {
		return nil
	}
	timestring := FmtTime(*t)
	return &timestring
}

func ParseTime(str string) (time.Time, error) {
	var firstErr error
	for _, layout := range parseFormats {
		t, err := time.Parse(layout, str)
		if err == nil {
			return t, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return time.Time{}, firstErr
}

// ParseTimePointer converts an optional string back to time.Time
func ParseTimePointer(str *string) (*time.Time, error) {
	if str == nil {
		return nil, nil
	}
	timeObject, err := ParseTime(*str)
	if err != nil {
		return nil, err
	}
	return &timeObject, nil
}

// RelativeTime returns the duration between the given time and now
func RelativeTime(t time.Time) time.Duration {
	return time.Since(t)
}

// HumanReadableDuration expresses a duration as a readable string
func HumanReadableDuration(dur time.Duration) string {

	rawSeconds := int(dur.Seconds())

	times := []struct {
		Name        string
		WholeCount  int
		RelativeMod int // remainder after division by next larger unit
	}{
		{"day", rawSeconds / (24 * 60 * 60), 0},
		{"hour", rawSeconds / (60 * 60), (rawSeconds % (24 * 60 * 60)) / (60 * 60)},
		{"minute", rawSeconds / 60, (rawSeconds % (60 * 60)) / 60},
		{"second", rawSeconds, rawSeconds % 60},
	}

	// express duration in largest unit with 1 or more complete units,
	// with any remainder expressed in the next smaller unit
	for idx, t := range times {
		if t.WholeCount > 0 {
			wholePlural := "s"
			if t.WholeCount == 1 {
				wholePlural = ""
			}
			outputString := fmt.Sprintf("%d %s", t.WholeCount, t.Name+wholePlural)
			if idx < len(times)-1 {
				nextSmaller := times[idx+1]
				if nextSmaller.RelativeMod > 0 {
					nextPlural := "s"
					if nextSmaller.RelativeMod == 1 {
						wholePlural = ""
					}
					outputString += " and " + fmt.Sprintf("%d %s", nextSmaller.RelativeMod, nextSmaller.Name+nextPlural)
				}
			}
			outputString += " ago"
			return outputString
		}
	}

	return "0 seconds ago"
}
