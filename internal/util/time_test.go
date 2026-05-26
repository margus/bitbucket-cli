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
