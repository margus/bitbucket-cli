package util

import (
	"strings"
	"testing"
)

func TestUcfirst(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"a", "A"},
		{"ab", "Ab"},
		{"Hello", "Hello"},
		{"hello world", "Hello world"},
		{"über", "Über"},
	}
	for _, tt := range tests {
		if got := ucfirst(tt.in); got != tt.want {
			t.Errorf("ucfirst(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestColorize(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	got := colorize("red", "hi")
	if !strings.Contains(got, "\033[0;31m") || !strings.Contains(got, "\033[0m") || !strings.Contains(got, "hi") {
		t.Errorf("colorize() = %q, want ANSI-wrapped text", got)
	}

	t.Setenv("NO_COLOR", "1")
	if got := colorize("red", "hi"); got != "hi" {
		t.Errorf("colorize() with NO_COLOR set = %q, want %q", got, "hi")
	}
}

func TestColorizeUnknownColor(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	// An unknown color name falls back to white rather than panicking.
	got := colorize("not-a-real-color", "hi")
	if !strings.Contains(got, "hi") {
		t.Errorf("colorize(unknown) = %q, want it to still contain the payload", got)
	}
}

func TestDisabled(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	if disabled() {
		t.Error("disabled() = true with NO_COLOR unset, want false")
	}
	t.Setenv("NO_COLOR", "anything")
	if !disabled() {
		t.Error("disabled() = false with NO_COLOR set, want true")
	}
}
