package util

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout temporarily redirects os.Stdout, runs fn, and returns
// what was written. NO_COLOR is set so the assertions can be plain text.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	t.Setenv("NO_COLOR", "1")

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()
	w.Close()
	os.Stdout = orig
	<-done
	return buf.String()
}

func TestOString(t *testing.T) {
	got := captureStdout(t, func() { O("hello", "green") })
	if !strings.Contains(got, "hello") {
		t.Errorf("O() output = %q, want it to contain hello", got)
	}
}

func TestOMap(t *testing.T) {
	got := captureStdout(t, func() {
		O(map[string]string{"branch": "main", "user": "alice"}, "yellow")
	})
	// Map keys are Title-cased before printing.
	if !strings.Contains(got, "Branch:") || !strings.Contains(got, "main") {
		t.Errorf("O(map) output = %q, want Branch: and main", got)
	}
	if !strings.Contains(got, "User:") || !strings.Contains(got, "alice") {
		t.Errorf("O(map) output = %q, want User: and alice", got)
	}
}

func TestOSliceOfMaps(t *testing.T) {
	got := captureStdout(t, func() {
		O([]map[string]string{
			{"id": "1"},
			{"id": "2"},
		}, "yellow")
	})
	if !strings.Contains(got, "1") || !strings.Contains(got, "2") {
		t.Errorf("O(slice) output = %q, want both entries", got)
	}
}

func TestORaw(t *testing.T) {
	got := captureStdout(t, func() { ORaw(".", "yellow") })
	if got != "." {
		// NO_COLOR strips escape codes, so output is the bare payload
		// with no trailing newline.
		t.Errorf("ORaw() = %q, want %q", got, ".")
	}
}
