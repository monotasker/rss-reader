package feed

import (
	"testing"
	"time"
)

func TestParseRSSTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "rfc1123z",
			input: "Sat, 25 Jul 2026 10:00:00 -0400",
			want:  time.Date(2026, 7, 25, 10, 0, 0, 0, time.FixedZone("", -4*3600)),
		},
		{
			name:  "rfc1123",
			input: "Sat, 25 Jul 2026 10:00:00 GMT",
			want:  time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC),
		},
		{
			name:  "rfc3339",
			input: "2026-07-25T10:00:00Z",
			want:  time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC),
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "garbage",
			input:   "not a date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRSSTime(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseRSSTime(%q) succeeded, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseRSSTime(%q): %v", tt.input, err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("parseRSSTime(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
