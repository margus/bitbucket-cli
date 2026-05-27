package util

import (
	"strings"
	"testing"
	"time"
)

func TestFormatRelativeTimestamp(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		input   string
		want    string
		wantSub string
	}{
		{"empty", "", "", ""},
		{"minutes ago", now.Add(-5 * time.Minute).Format(time.RFC3339), "", "minutes ago"},
		{"one minute ago", now.Add(-90 * time.Second).Format(time.RFC3339), "1 minute ago", ""},
		{"hours ago", now.Add(-3 * time.Hour).Format(time.RFC3339), "3 hours ago", ""},
		{"one hour ago", now.Add(-90 * time.Minute).Format(time.RFC3339), "1 hour ago", ""},
		{"days ago", now.Add(-3 * 24 * time.Hour).Format(time.RFC3339), "3 days ago", ""},
		{"older than 7 days", now.Add(-30 * 24 * time.Hour).Format(time.RFC3339), "", ""},
		{"future date", now.Add(48 * time.Hour).Format(time.RFC3339), "", ""},
		{"unparseable", "not a date", "not a date", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRelativeTimestamp(tt.input)
			if tt.want != "" && got != tt.want {
				t.Errorf("FormatRelativeTimestamp(%q) = %q, want %q", tt.input, got, tt.want)
			}
			if tt.wantSub != "" && !strings.Contains(got, tt.wantSub) {
				t.Errorf("FormatRelativeTimestamp(%q) = %q, want substring %q", tt.input, got, tt.wantSub)
			}
		})
	}
}

// FuzzFormatRelativeTimestamp throws arbitrary strings at the
// timestamp formatter to catch panics on malformed input. The
// function should always return some string, never crash.
func FuzzFormatRelativeTimestamp(f *testing.F) {
	for _, seed := range []string{
		"2026-05-27T12:34:56Z",
		"2026-05-27T12:34:56+02:00",
		"not a date",
		"",
		"0000-00-00T00:00:00Z",
		"9999-12-31T23:59:59Z",
		"2026-05-27",
		strings.Repeat("a", 1024),
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, in string) {
		// We don't care what comes out; only that nothing panics and
		// the function terminates.
		_ = FormatRelativeTimestamp(in)
	})
}

func TestPluralAgo(t *testing.T) {
	tests := []struct {
		n    int
		unit string
		want string
	}{
		{1, "day", "1 day ago"},
		{2, "day", "2 days ago"},
		{1, "hour", "1 hour ago"},
		{5, "minute", "5 minutes ago"},
	}
	for _, tt := range tests {
		if got := pluralAgo(tt.n, tt.unit); got != tt.want {
			t.Errorf("pluralAgo(%d, %q) = %q, want %q", tt.n, tt.unit, got, tt.want)
		}
	}
}
