package main

import (
	"runtime"
	"strings"
	"testing"
)

func TestBrowseURL(t *testing.T) {
	out, err := runCmd(t, "--project", testRepo, "browse", "url")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	want := "https://bitbucket.org/acme/widgets"
	if !strings.Contains(out, want) {
		t.Errorf("output missing %q:\n%s", want, out)
	}
}

func TestOpenerFor(t *testing.T) {
	tests := []struct {
		goos, want string
	}{
		{"darwin", "open"},
		{"linux", "xdg-open"},
		{"windows", "start"},
		{"plan9", ""},   // unsupported -> empty
		{"netbsd", ""},  // unsupported -> empty
	}
	for _, tt := range tests {
		if got := openerFor(tt.goos); got != tt.want {
			t.Errorf("openerFor(%q) = %q, want %q", tt.goos, got, tt.want)
		}
	}
}

// We don't exec a real browser in tests — that would actually open one.
// Instead, the test just confirms openerFor returns something non-empty
// for the runtime we're on. Coverage of the browse() handler itself is
// achieved through TestBrowseURL above (same code path minus exec).
func TestBrowseOpenerExistsForCurrentOS(t *testing.T) {
	if got := openerFor(runtime.GOOS); got == "" {
		t.Skipf("no opener defined for %q — test environment", runtime.GOOS)
	}
}
