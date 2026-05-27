package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestRepoListExplicitWorkspace(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/repositories/acme") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"values":[{"slug":"widgets","name":"Widgets","language":"go","is_private":true},{"slug":"forms","name":"Forms","language":"php","is_private":false}]}`))
	}))

	out, err := runCmd(t, "repo", "list", "acme")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"widgets", "forms", "go", "php", "private", "public"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRepoListInferredFromProject(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We passed --project acme/widgets via runRepoCmd; the workspace
		// half should be "acme".
		if !strings.HasPrefix(r.URL.Path, "/repositories/acme") {
			t.Errorf("expected /repositories/acme prefix, got: %s", r.URL.Path)
		}
		w.Write([]byte(`{"values":[{"slug":"widgets","name":"Widgets"}]}`))
	}))

	out, err := runRepoCmd(t, "repo", "list")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "widgets") {
		t.Errorf("output missing slug:\n%s", out)
	}
}

func TestRepoListYAMLOutput(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"values":[{"slug":"widgets","name":"Widgets","language":"go"}]}`))
	}))

	out, err := runCmd(t, "--output", "yaml", "repo", "list", "acme")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "slug:") || !strings.Contains(out, "widgets") {
		t.Errorf("expected YAML output:\n%s", out)
	}
}
