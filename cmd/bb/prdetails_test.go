package main

import (
	"net/http"
	"strings"
	"testing"
)

const prCommentsPage = `{
  "values":[
    {"created_on":"2026-05-20T10:00:00Z","user":{"display_name":"Alice"},"content":{"raw":"general comment from alice"}},
    {"created_on":"2026-05-21T11:00:00Z","user":{"display_name":"Bob"},"content":{"raw":"inline note from bob"},"inline":{"path":"foo.go","to":42},"resolution":{}},
    {"created_on":"2026-05-22T12:00:00Z","user":{"display_name":"Carol"},"content":{"raw":"unresolved inline from carol"},"inline":{"path":"bar.go","to":5}}
  ]
}`

func TestPrDetailsShow(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/pullrequests/7/comments") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(prCommentsPage))
	}))

	out, err := runRepoCmd(t, "pr", "show", "7")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"general comment from alice", "inline note from bob", "unresolved inline from carol", "foo.go:42", "bar.go:5"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrDetailsShowUnresolvedFilter(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(prCommentsPage))
	}))

	out, err := runRepoCmd(t, "pr", "show", "7", "true")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	// General comments are never resolvable per the API — they stay.
	if !strings.Contains(out, "general comment from alice") {
		t.Errorf("general comment should remain when filtering inline only:\n%s", out)
	}
	// Carol's inline (no resolution key) stays.
	if !strings.Contains(out, "unresolved inline from carol") {
		t.Errorf("unresolved inline missing:\n%s", out)
	}
	// Bob's inline (has resolution: {}) gets filtered out.
	if strings.Contains(out, "inline note from bob") {
		t.Errorf("resolved inline should be filtered out:\n%s", out)
	}
}

func TestPrDetailsShowDeletedComments(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"values":[{"created_on":"2026-05-20T10:00:00Z","user":{"display_name":"X"},"content":{"raw":"orig"},"deleted":true}]}`))
	}))

	out, err := runRepoCmd(t, "pr", "show", "7")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "[DELETED]") {
		t.Errorf("expected [DELETED] marker:\n%s", out)
	}
	if strings.Contains(out, "orig") {
		t.Errorf("deleted comment body should not render:\n%s", out)
	}
}

func TestPrDetailsShowEmpty(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"values":[]}`))
	}))

	out, err := runRepoCmd(t, "pr", "show", "7")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "No general comments") || !strings.Contains(out, "No inline comments") {
		t.Errorf("expected 'no comments' messages:\n%s", out)
	}
}
