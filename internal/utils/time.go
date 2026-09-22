package utils

import (
	"fmt"
	"time"
)

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
