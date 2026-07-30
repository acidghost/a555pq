package timespan

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		// plain units
		{in: "2y", want: 2 * 365 * 24 * time.Hour},
		{in: "6mo", want: 6 * 30 * 24 * time.Hour},
		{in: "7d", want: 7 * 24 * time.Hour},
		{in: "10h", want: 10 * time.Hour},
		{in: "1w", want: 7 * 24 * time.Hour},
		{in: "30m", want: 30 * time.Minute},
		{in: "60s", want: 60 * time.Second},
		{in: "500ms", want: 500 * time.Millisecond},
		// combinations, including calendar-unit approximations
		{in: "1y6mo", want: 365*24*time.Hour + 6*30*24*time.Hour},
		{in: "1w2d", want: 9 * 24 * time.Hour},
		{in: "1d2h3m", want: 24*time.Hour + 2*time.Hour + 3*time.Minute},
		{in: "1d 2h", want: 24*time.Hour + 2*time.Hour}, // whitespace between pairs
		// fractional values
		{in: "1.5d", want: 36 * time.Hour},
		{in: "0.5h", want: 30 * time.Minute},
		// m vs ms vs mo disambiguation
		{in: "1m", want: time.Minute},
		{in: "1ms", want: time.Millisecond},
		{in: "1mo", want: 30 * 24 * time.Hour},
		// case insensitivity
		{in: "7D", want: 7 * 24 * time.Hour},
		{in: "10H30M", want: 10*time.Hour + 30*time.Minute},
		// zero / empty
		{in: "", want: 0},
		{in: "0s", want: 0},
		// errors
		{in: "0", wantErr: true},      // number without unit
		{in: "5", wantErr: true},      // number without unit
		{in: "5x", wantErr: true},     // unknown unit
		{in: "abc", wantErr: true},    // not a number
		{in: "-5d", wantErr: true},    // negatives not allowed
		{in: "5d5", wantErr: true},    // trailing number without unit
		{in: "1.2.3d", wantErr: true}, // multiple decimal points
		{in: "d", wantErr: true},      // unit without number
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := Parse(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) = %v, want error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
