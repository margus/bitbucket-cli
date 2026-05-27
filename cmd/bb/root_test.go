package main

import (
	"os"
	"strings"
	"testing"
)

// TestRootCommandsRegistered exercises the wiring without going through
// the cobra flag parser. Cobra bool flags (--help, --version) stick
// across Execute() calls, so driving them via SetArgs+Execute poisons
// subsequent tests. Verifying registration directly is both faster and
// more reliable.
func TestRootCommandsRegistered(t *testing.T) {
	registered := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		registered[c.Name()] = true
	}
	for _, want := range []string{"auth", "branch", "browse", "env", "pipeline", "pr", "pr-details", "upgrade", "completion"} {
		if !registered[want] {
			t.Errorf("rootCmd missing subcommand: %s\nregistered: %v", want, registered)
		}
	}
}

func TestVersionString(t *testing.T) {
	saved := version
	savedCommit := commit
	savedDate := date
	version = "v1.2.3"
	commit = "abc1234"
	date = "2026-05-27"
	t.Cleanup(func() { version = saved; commit = savedCommit; date = savedDate })

	got := versionString()
	for _, want := range []string{"v1.2.3", "abc1234", "2026-05-27"} {
		if !strings.Contains(got, want) {
			t.Errorf("versionString() missing %q: %q", want, got)
		}
	}
}

func TestRequireGitRejectsWithoutProject(t *testing.T) {
	tmp := t.TempDir()
	saved, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(saved) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	withSandboxedConfig(t)

	out, err := runCmd(t, "branch", "list")
	if err == nil {
		t.Fatal("expected error when not in a git repo and no --project")
	}
	if !strings.Contains(out, "No git repository found") {
		t.Errorf("missing git-required hint:\n%s", out)
	}
}

func TestUnknownAction(t *testing.T) {
	withSandboxedConfig(t)
	out, err := runCmd(t, "bogus")
	if err == nil {
		t.Errorf("expected error for unknown action, got nil. Output: %s", out)
	}
}
