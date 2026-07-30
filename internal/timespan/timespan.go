// Package timespan parses human-friendly durations such as "7d", "10h", or
// "1w2d3h" into [time.Duration]. Unlike [time.ParseDuration], it understands
// days ("d") and weeks ("w"), which are needed for release-age filtering.
package timespan

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var unitMap = map[string]time.Duration{
	"y":  365 * 24 * time.Hour, // year  (approximated as 365 days)
	"mo": 30 * 24 * time.Hour,  // month (approximated as 30 days)
	"w":  7 * 24 * time.Hour,
	"d":  24 * time.Hour,
	"h":  time.Hour,
	"m":  time.Minute,
	"s":  time.Second,
	"ms": time.Millisecond,
}

// Parse parses a timespan into a [time.Duration].
//
// A timespan is one or more value-unit pairs concatenated together, for
// example "7d", "10h", "1w2d", or "1d6h30m". Whitespace between pairs is
// allowed. Supported units (case-insensitive):
//
//	y  years   (approximated as 365 days)
//	mo months  (approximated as 30 days)
//	w  weeks
//	d  days
//	h  hours
//	m  minutes
//	s  seconds
//	ms milliseconds
//
// Years and months are approximated as fixed durations since a [time.Duration]
// cannot represent calendar units. Values may be fractional (e.g. "1.5d"). The
// empty string parses to a zero duration. A negative sign is not allowed.
func Parse(s string) (time.Duration, error) {
	orig := s
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	var total time.Duration
	for s != "" {
		s = strings.TrimSpace(s)

		i := 0
		dots := 0
		for i < len(s) {
			c := s[i]
			if c >= '0' && c <= '9' {
				i++
				continue
			}
			if c == '.' {
				dots++
				if dots > 1 {
					return 0, fmt.Errorf("timespan: invalid number %q in %q", orig, orig)
				}
				i++
				continue
			}
			break
		}
		if i == 0 {
			return 0, fmt.Errorf("timespan: expected number in %q", orig)
		}
		numStr := s[:i]

		// unit: up to two ASCII letters, matched greedily so "ms" wins over "m".
		j := i
		for j < len(s) && j < i+2 {
			c := s[j]
			if !isLetter(c) {
				break
			}
			j++
		}
		unitStr := strings.ToLower(s[i:j])
		unit, ok := unitMap[unitStr]
		if !ok {
			return 0, fmt.Errorf("timespan: unknown unit %q in %q", unitStr, orig)
		}

		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("timespan: invalid number %q in %q: %w", numStr, orig, err)
		}
		total += time.Duration(num * float64(unit))
		s = s[j:]
	}

	return total, nil
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
