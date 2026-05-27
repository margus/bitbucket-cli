package util

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestRepoPathFromProjectURL(t *testing.T) {
	tests := []struct {
		name    string
		project string
		want    string
		wantErr bool
	}{
		{"owner/repo", "acme/widgets", "acme/widgets", false},
		{"https url", "https://bitbucket.org/acme/widgets", "acme/widgets", false},
		{"https url with .git", "https://bitbucket.org/acme/widgets.git", "acme/widgets", false},
		{"https url with trailing slash", "https://bitbucket.org/acme/widgets/", "acme/widgets", false},
		{"ssh url", "git@bitbucket.org:acme/widgets.git", "acme/widgets", false},
		{"invalid", "not a url at all !!", "", true},
		{"empty", "", "", true}, // falls through to git, which won't have a bitbucket remote in CI
	}

	// Save and restore so tests don't leak state.
	saved := ProjectURL
	defer func() { ProjectURL = saved }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ProjectURL = tt.project
			got, err := RepoPath()
			if tt.wantErr {
				if err == nil {
					t.Errorf("RepoPath() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Errorf("RepoPath() error = %v, want %q", err, tt.want)
				return
			}
			if got != tt.want {
				t.Errorf("RepoPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

// withGitRepo cd's into a fresh tempdir, runs `git init`, and creates
// an initial commit on a branch named `feature/test`. Returns the
// tempdir path. Restores cwd on cleanup.
func withGitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	tmp := t.TempDir()
	saved, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(saved) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	runGit("init", "-b", "feature/test", ".")
	if err := os.WriteFile("hello", []byte("hi"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit("add", "hello")
	runGit("commit", "-m", "initial")
	return tmp
}

func TestHasGitDirInsideRepo(t *testing.T) {
	withGitRepo(t)
	if !HasGitDir() {
		t.Error("HasGitDir() = false inside a git repo")
	}
}

func TestHasGitDirOutsideRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	tmp := t.TempDir()
	saved, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(saved) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if HasGitDir() {
		t.Error("HasGitDir() = true outside any git repo")
	}
}

func TestCurrentBranch(t *testing.T) {
	withGitRepo(t)
	got, err := CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if got != "feature/test" {
		t.Errorf("CurrentBranch() = %q, want %q", got, "feature/test")
	}
}

// FuzzRepoPathFromProjectURL throws arbitrary strings at the
// --project parser. None should cause a panic — only "looks like a
// repo path" results, or an error. The parser is purely regex-based
// so this is mostly a cheap insurance policy against future edits.
func FuzzRepoPathFromProjectURL(f *testing.F) {
	for _, seed := range []string{
		"acme/widgets",
		"https://bitbucket.org/acme/widgets",
		"https://bitbucket.org/acme/widgets.git",
		"git@bitbucket.org:acme/widgets.git",
		"",
		"!@#$%^&*()",
		"a/b/c/d",
		strings.Repeat("x", 1024),
		"acme//widgets",
		"git@bitbucket.org:",
	} {
		f.Add(seed)
	}
	saved := ProjectURL
	defer func() { ProjectURL = saved }()
	f.Fuzz(func(t *testing.T, in string) {
		ProjectURL = in
		// We don't care about the result here, only that we don't panic.
		_, _ = RepoPath()
	})
}

func TestCurrentBranchOutsideRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	tmp := t.TempDir()
	saved, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(saved) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if _, err := CurrentBranch(); err == nil {
		t.Error("CurrentBranch() outside a repo should return an error")
	}
}
