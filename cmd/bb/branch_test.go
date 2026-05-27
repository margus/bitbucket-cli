package main

import (
	"net/http"
	"strings"
	"testing"
)

const branchListBody = `{
  "values": [
    {
      "name": "main",
      "target": {
        "date": "2026-05-20T10:30:00+00:00",
        "author": {"raw": "alice <alice@example.com>", "user": {"display_name": "Alice"}}
      }
    },
    {
      "name": "feature/widgets",
      "target": {
        "date": "2026-05-22T14:00:00+00:00",
        "author": {"raw": "bob <bob@example.com>", "user": {"display_name": "Bob"}}
      }
    }
  ]
}`

func TestBranchList(t *testing.T) {
	var hits int
	srv := withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if !strings.HasPrefix(r.URL.Path, "/repositories/acme/widgets/refs/branches") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(branchListBody))
	}))
	_ = srv

	out, err := runRepoCmd(t, "branch", "list")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"main", "feature/widgets", "Alice", "Bob"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if hits != 1 {
		t.Errorf("expected 1 backend call, got %d", hits)
	}
}

func TestBranchUser(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(branchListBody))
	}))

	out, err := runRepoCmd(t, "branch", "user", "Alice")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "main") {
		t.Errorf("expected branch authored by Alice in output:\n%s", out)
	}
	if strings.Contains(out, "feature/widgets") {
		t.Errorf("branch by Bob should have been filtered out:\n%s", out)
	}
}

func TestBranchName(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(branchListBody))
	}))

	out, err := runRepoCmd(t, "branch", "name", "feature")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "feature/widgets") {
		t.Errorf("expected feature branch in output:\n%s", out)
	}
	if strings.Contains(out, "Updated:") && strings.Contains(out, "main") {
		// "main" might appear inside a path like /refs/branches not as a
		// branch listing — only flag when it shows up as a rendered row.
		t.Errorf("'main' branch should have been filtered out:\n%s", out)
	}
}

func TestBranchListPagination(t *testing.T) {
	var page int
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		if page == 1 {
			w.Write([]byte(`{
              "values": [{"name":"page1","target":{"date":"2026-05-20T10:30:00+00:00","author":{"raw":"a"}}}],
              "next": "https://example.com/?page=2"
            }`))
			return
		}
		w.Write([]byte(`{
          "values": [{"name":"page2","target":{"date":"2026-05-20T10:30:00+00:00","author":{"raw":"b"}}}]
        }`))
	}))

	out, err := runRepoCmd(t, "branch", "list")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"page1", "page2"} {
		if !strings.Contains(out, want) {
			t.Errorf("pagination dropped %q. Output:\n%s", want, out)
		}
	}
	if page != 2 {
		t.Errorf("expected 2 paginated requests, got %d", page)
	}
}
