package main

import (
	"reflect"
	"testing"
)

func TestParseGlobalFlags(t *testing.T) {
	tests := []struct {
		name      string
		argv      []string
		wantFlags globalFlags
		wantRest  []string
	}{
		{
			name:      "empty",
			argv:      nil,
			wantFlags: globalFlags{},
			wantRest:  []string{},
		},
		{
			name:      "no flags",
			argv:      []string{"pr", "list"},
			wantFlags: globalFlags{},
			wantRest:  []string{"pr", "list"},
		},
		{
			name:      "project flag",
			argv:      []string{"--project", "acme/widgets", "pr", "list"},
			wantFlags: globalFlags{Project: "acme/widgets"},
			wantRest:  []string{"pr", "list"},
		},
		{
			name:      "interactive long",
			argv:      []string{"--interactive", "pr", "create", "main"},
			wantFlags: globalFlags{Interactive: true},
			wantRest:  []string{"pr", "create", "main"},
		},
		{
			name:      "interactive short",
			argv:      []string{"-i", "pr", "create", "main"},
			wantFlags: globalFlags{Interactive: true},
			wantRest:  []string{"pr", "create", "main"},
		},
		{
			name:      "title and description",
			argv:      []string{"--title", "hi", "--description", "a desc", "pr", "create", "main"},
			wantFlags: globalFlags{Title: "hi", Description: "a desc"},
			wantRest:  []string{"pr", "create", "main"},
		},
		{
			name:      "all flags",
			argv:      []string{"--project", "a/b", "-i", "--title", "t", "--description", "d", "pr", "create", "main"},
			wantFlags: globalFlags{Project: "a/b", Title: "t", Description: "d", Interactive: true},
			wantRest:  []string{"pr", "create", "main"},
		},
		{
			name:      "flag value missing at end",
			argv:      []string{"--project"},
			wantFlags: globalFlags{},
			wantRest:  []string{},
		},
		{
			name:      "flags interleaved",
			argv:      []string{"pr", "--project", "a/b", "create", "main"},
			wantFlags: globalFlags{Project: "a/b"},
			wantRest:  []string{"pr", "create", "main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gf, rest := parseGlobalFlags(tt.argv)
			if !reflect.DeepEqual(gf, tt.wantFlags) {
				t.Errorf("flags = %+v, want %+v", gf, tt.wantFlags)
			}
			if !reflect.DeepEqual(rest, tt.wantRest) {
				t.Errorf("rest = %v, want %v", rest, tt.wantRest)
			}
		})
	}
}

func TestIsHelpVersionAutocomplete(t *testing.T) {
	for _, s := range []string{"help", "--help", "-h"} {
		if !isHelp(s) {
			t.Errorf("isHelp(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"version", "--version", "-v"} {
		if !isVersion(s) {
			t.Errorf("isVersion(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"autocomplete", "--autocomplete"} {
		if !isAutocomplete(s) {
			t.Errorf("isAutocomplete(%q) = false, want true", s)
		}
	}
	if isHelp("pr") || isVersion("pr") || isAutocomplete("pr") {
		t.Error("non-static command incorrectly matched a static command predicate")
	}
}
