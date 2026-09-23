package utils_test

import (
	"testing"
	"time"

	"github.com/monotasker/rss-reader/internal/utils"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want time.Time
	}{
		{
			name: "rfc3339",
			in:   "2026-09-23T01:19:47Z",
			want: time.Date(2026, 9, 23, 1, 19, 47, 0, time.UTC),
		},
		{
			name: "rfc3339 nano",
			in:   "2026-09-23T01:19:47.964855Z",
			want: time.Date(2026, 9, 23, 1, 19, 47, 964855000, time.UTC),
		},
		{
			name: "sqlite driver legacy",
			in:   "2026-09-23 01:19:47.964855+00:00",
			want: time.Date(2026, 9, 23, 1, 19, 47, 964855000, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.ParseTime(tt.in)
			if err != nil {
				t.Fatalf("ParseTime(%q): %v", tt.in, err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("ParseTime(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseTimeInvalid(t *testing.T) {
	if _, err := utils.ParseTime("not-a-time"); err == nil {
		t.Fatal("expected error for invalid time")
	}
}

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
