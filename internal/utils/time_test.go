package utils_test

import (
	"testing"
	"time"

	"github.com/monotasker/rss-reader/internal/utils"
)

func TestHumanReadableDuration(t *testing.T) {
	tests := []struct {
		name     string
		dur      time.Duration
		expected string
	}{
		{"days only", 36*time.Hour*24 + 12*time.Minute, "36 days ago"},
		{"days/hours", 36*time.Hour*24 + 16*time.Hour + 24*time.Second, "36 days and 16 hours ago"},
		{"hours/minutes", 16*time.Hour + 24*time.Minute, "16 hours and 24 minutes ago"},
		{"minutes/seconds", 1*time.Minute + 20*time.Second, "1 minute and 20 seconds ago"},
		{"seconds only", 5 * time.Second, "5 seconds ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := utils.HumanReadableDuration(tt.dur)
			if actual != tt.expected {
				t.Errorf("HumanReadableDuration(%v) = %v; expected %v",
					tt.dur, actual, tt.expected)
			}
		})
	}
}
