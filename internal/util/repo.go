package util

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// ProjectURL is the value of --project, set in main before action dispatch.
var ProjectURL string

var (
	ownerRepoRe = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.-]*[a-zA-Z0-9_]/[a-zA-Z0-9_][a-zA-Z0-9_.-]*[a-zA-Z0-9_]$`)
	httpRe      = regexp.MustCompile(`^https?://bitbucket\.org/(.+?)/?(?:\.git)?/?$`)
	sshRe       = regexp.MustCompile(`^git@bitbucket\.org:(.+?)\.git$`)
	remoteRe    = regexp.MustCompile(`.*bitbucket\.org[:,/](.+?)\.git`)
)

// RepoPath returns "owner/repo" from --project flag or git remote.origin.url.
func RepoPath() (string, error) {
	if ProjectURL != "" {
		if ownerRepoRe.MatchString(ProjectURL) {
			return ProjectURL, nil
		}
		for _, re := range []*regexp.Regexp{httpRe, sshRe} {
			if m := re.FindStringSubmatch(ProjectURL); m != nil {
				return strings.TrimRight(m[1], "/"), nil
			}
		}
		return "", fmt.Errorf(`invalid repository format. Expected: "owner/repo" or "https://bitbucket.org/owner/repo"`)
	}

	out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", fmt.Errorf("cannot get repository info. Are you sure this is a bitbucket repository?")
	}
	m := remoteRe.FindStringSubmatch(strings.TrimSpace(string(out)))
	if m == nil {
		return "", fmt.Errorf("cannot get repository info. Are you sure this is a bitbucket repository?")
	}
	return m[1], nil
}

// HasGitDir reports whether the current directory is inside a git repo.
func HasGitDir() bool {
	out, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// CurrentBranch returns the current git branch name (HEAD).
func CurrentBranch() (string, error) {
	out, err := exec.Command("git", "symbolic-ref", "--short", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("getting current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
