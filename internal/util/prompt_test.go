package util

import (
	"io"
	"os"
	"testing"
)

func TestPromptCallsPromptFn(t *testing.T) {
	saved := PromptFn
	PromptFn = func(q, def string) string { return "stubbed" }
	t.Cleanup(func() { PromptFn = saved })

	if got := Prompt("anything", ""); got != "stubbed" {
		t.Errorf("Prompt() = %q, want %q", got, "stubbed")
	}
}

// readLineFromStdin reads from os.Stdin. To test it we have to redirect
// stdin via a pipe.
func TestReadLineFromStdin(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()

	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })

	go func() {
		_, _ = io.WriteString(w, "alice\n")
		w.Close()
	}()

	got := readLineFromStdin("Username:", "")
	if got != "alice" {
		t.Errorf("readLineFromStdin = %q, want %q", got, "alice")
	}
}

func TestReadLineFromStdinReturnsDefaultOnEmpty(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()

	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })

	go func() {
		_, _ = io.WriteString(w, "\n")
		w.Close()
	}()

	got := readLineFromStdin("Q:", "the-default")
	if got != "the-default" {
		t.Errorf("readLineFromStdin = %q, want default %q", got, "the-default")
	}
}

func TestReadLineFromStdinReturnsDefaultOnEOF(t *testing.T) {
	r, w, _ := os.Pipe()
	defer r.Close()
	w.Close() // immediate EOF

	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })

	got := readLineFromStdin("Q:", "fallback")
	if got != "fallback" {
		t.Errorf("readLineFromStdin = %q, want fallback %q", got, "fallback")
	}
}
