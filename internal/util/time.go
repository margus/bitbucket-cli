package util

import (
	"fmt"
	"time"
)

// FormatRelativeTimestamp formats an ISO-8601 timestamp: "X ago" for
// past dates within 7 days, "Jan 02, 2006" for older or future dates.
func FormatRelativeTimestamp(s string) string {
	if s == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try without timezone fraction
		if t2, err2 := time.Parse("2006-01-02T15:04:05Z07:00", s); err2 == nil {
			t = t2
		} else {
			return s
		}
	}

	now := time.Now()
	if t.After(now) {
		return t.Format("Jan 02, 2006")
	}

	diff := now.Sub(t)
	days := int(diff.Hours() / 24)
	switch {
	case days > 7:
		return t.Format("Jan 02, 2006")
	case days > 0:
		return pluralAgo(days, "day")
	case int(diff.Hours()) > 0:
		return pluralAgo(int(diff.Hours()), "hour")
	default:
		return pluralAgo(int(diff.Minutes()), "minute")
	}
}

func pluralAgo(n int, unit string) string {
	if n > 1 {
		return fmt.Sprintf("%d %ss ago", n, unit)
	}
	return fmt.Sprintf("%d %s ago", n, unit)
}
