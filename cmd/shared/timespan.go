package shared

import (
	"time"

	"github.com/acidghost/a555pq/internal/timespan"
)

// TimespanValue is a [pflag.Value] that parses human-friendly timespans such as
// "7d", "10h", or "1w2d" into a [time.Duration] using [timespan.Parse].
type TimespanValue struct {
	duration *time.Duration
}

// NewTimespanValue wraps the given duration pointer so a flag binding can update
// it. The flag's zero value (an unset flag) leaves the duration at zero.
func NewTimespanValue(d *time.Duration) *TimespanValue {
	return &TimespanValue{duration: d}
}

func (t *TimespanValue) Set(s string) error {
	parsed, err := timespan.Parse(s)
	if err != nil {
		return err
	}
	*t.duration = parsed
	return nil
}

func (t *TimespanValue) String() string {
	if t.duration == nil {
		return "0s"
	}
	return t.duration.String()
}

// Type is the name shown in flag usage and completions.
func (t *TimespanValue) Type() string { return "timespan" }
