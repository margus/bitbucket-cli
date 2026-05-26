package util

import "testing"

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
