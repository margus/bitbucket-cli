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

// captureStderr is the stderr sibling of captureStdout.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	t.Setenv("NO_COLOR", "1")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stderr
	os.Stderr = w
	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()
	fn()
	w.Close()
	os.Stderr = orig
	<-done
	return buf.String()
}

func TestErrorln(t *testing.T) {
	got := captureStderr(t, func() { Errorln("boom: %s", "bang") })
	if !strings.Contains(got, "boom: bang") {
		t.Errorf("Errorln() = %q, want it to contain 'boom: bang'", got)
	}
}

func TestOWithNestedSlice(t *testing.T) {
	out := captureStdout(t, func() {
		O(map[string]any{
			"reviewers": []string{"alice", "bob"},
		}, "yellow")
	})
	if !strings.Contains(out, "alice") || !strings.Contains(out, "bob") {
		t.Errorf("nested slice not rendered:\n%s", out)
	}
}

func TestOWithNestedMap(t *testing.T) {
	out := captureStdout(t, func() {
		O(map[string]any{
			"author": map[string]string{"name": "alice"},
		}, "yellow")
	})
	if !strings.Contains(out, "alice") {
		t.Errorf("nested map not rendered:\n%s", out)
	}
}

func TestOWithNumberAndBool(t *testing.T) {
	out := captureStdout(t, func() {
		O(42, "yellow")
	})
	if !strings.Contains(out, "42") {
		t.Errorf("int not rendered: %q", out)
	}
	out2 := captureStdout(t, func() {
		O(true, "yellow")
	})
	if !strings.Contains(out2, "true") {
		t.Errorf("bool not rendered: %q", out2)
	}
}

func TestONil(t *testing.T) {
	out := captureStdout(t, func() {
		O(nil, "yellow")
	})
	// Nil renders as an empty line (just the newline).
	if out != "\n" {
		t.Errorf("nil rendered as %q, want a single newline", out)
	}
}
