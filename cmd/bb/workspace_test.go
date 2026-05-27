package main

import (
	"net/http"
	"strings"
	"testing"
)

// Real shape returned by /2.0/user/workspaces, captured from the live
// Bitbucket Cloud API on 2026-05-27.
const userWorkspacesBody = `{
  "values": [
    {"type":"workspace_access","administrator":false,"workspace":{"type":"workspace_base","uuid":"{30951101-46a0-41c3-844d-3701faa05d79}","slug":"acme","links":{"self":{"href":"https://api.bitbucket.org/2.0/workspaces/acme"}}}},
    {"type":"workspace_access","administrator":true,"workspace":{"type":"workspace_base","uuid":"{57e59e5c-abc1-490a-a31b-9c94f543eba4}","slug":"widgets","links":{"self":{"href":"https://api.bitbucket.org/2.0/workspaces/widgets"}}}}
  ],
  "pagelen":100,"size":2,"page":1
}`

func TestWorkspaceList(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CHANGE-2770: must use /user/workspaces; the old /workspaces
		// and /user/permissions/workspaces both 410 Gone.
		if r.URL.Path != "/user/workspaces" {
			t.Errorf("unexpected path: %s (expected /user/workspaces)", r.URL.Path)
		}
		w.Write([]byte(userWorkspacesBody))
	}))

	out, err := runCmd(t, "workspace", "list")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	// Slug and UUID for each, plus administrator status.
	for _, want := range []string{"acme", "widgets", "30951101", "57e59e5c", "Yes", "No"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestWorkspaceListJSONOutput(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(userWorkspacesBody))
	}))

	out, err := runCmd(t, "--output", "json", "workspace", "list")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	out = strings.TrimSpace(out)
	if !strings.HasPrefix(out, "[") || !strings.HasSuffix(out, "]") {
		t.Errorf("expected JSON array, got:\n%s", out)
	}
	if !strings.Contains(out, `"slug": "acme"`) || !strings.Contains(out, `"administrator": true`) {
		t.Errorf("JSON output missing expected fields:\n%s", out)
	}
}

func TestWorkspaceCmdDefaultsToList(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"values":[{"administrator":false,"workspace":{"slug":"a","uuid":"{a}"}}]}`))
	}))

	out, err := runCmd(t, "workspace")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "a") {
		t.Errorf("default subcommand didn't run list:\n%s", out)
	}
}

func TestWorkspaceListPagination(t *testing.T) {
	var page int
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		if page == 1 {
			w.Write([]byte(`{"values":[{"administrator":true,"workspace":{"slug":"p1","uuid":"{p1}"}}],"next":"x?page=2"}`))
			return
		}
		w.Write([]byte(`{"values":[{"administrator":false,"workspace":{"slug":"p2","uuid":"{p2}"}}]}`))
	}))

	out, err := runCmd(t, "workspace", "list")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"p1", "p2"} {
		if !strings.Contains(out, want) {
			t.Errorf("pagination dropped %q:\n%s", want, out)
		}
	}
}
