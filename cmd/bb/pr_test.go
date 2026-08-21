package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/margus/bitbucket-cli/internal/util"
)

// prRouter dispatches by URL path prefix, longest match first so
// `/foo/7` wins over `/foo`. Query string is ignored — only the path
// is matched.
func prRouter(routes map[string]http.HandlerFunc) http.Handler {
	prefixes := make([]string, 0, len(routes))
	for k := range routes {
		prefixes = append(prefixes, k)
	}
	sort.Slice(prefixes, func(i, j int) bool { return len(prefixes[i]) > len(prefixes[j]) })
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, p := range prefixes {
			if strings.HasPrefix(r.URL.Path, p) {
				routes[p](w, r)
				return
			}
		}
		http.NotFound(w, r)
	})
}

const prListBody = `{"values":[{
  "id": 7,
  "author":{"nickname":"alice"},
  "source":{"branch":{"name":"feature/foo"}},
  "destination":{"branch":{"name":"main"}},
  "links":{"html":{"href":"https://bitbucket.org/acme/widgets/pull-requests/7"}}
}]}`

const prDetailBody = `{
  "reviewers":[{"display_name":"Bob"}],
  "participants":[{"state":"approved","user":{"display_name":"Carol"}}]
}`

func TestPrList(t *testing.T) {
	withTestServer(t, prRouter(map[string]http.HandlerFunc{
		"/repositories/acme/widgets/pullrequests": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(prListBody))
		},
		"/repositories/acme/widgets/pullrequests/7": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(prDetailBody))
		},
	}))

	out, err := runRepoCmd(t, "pr", "list")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"feature/foo", "main", "alice", "Bob", "Carol -> approved"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrListFilterByDestination(t *testing.T) {
	withTestServer(t, prRouter(map[string]http.HandlerFunc{
		"/repositories/acme/widgets/pullrequests": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(prListBody))
		},
		"/repositories/acme/widgets/pullrequests/7": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(prDetailBody))
		},
	}))

	out, err := runRepoCmd(t, "pr", "list", "develop") // PR is targeting main, not develop
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if strings.Contains(out, "feature/foo") {
		t.Errorf("PR targeting main should have been filtered out:\n%s", out)
	}
}

func TestPrDiff(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/pullrequests/7/diff") {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte("diff --git a/file b/file\n+added line"))
	}))

	out, err := runRepoCmd(t, "pr", "diff", "7")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "diff --git") {
		t.Errorf("expected raw diff in output:\n%s", out)
	}
}

func TestPrFiles(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"values":[{"new":{"path":"foo.go"}},{"new":{"path":"bar.md"}}]}`))
	}))

	out, err := runRepoCmd(t, "pr", "files", "7")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"foo.go", "bar.md"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrCommits(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"values":[{"summary":{"raw":"fix: tighten validation"}},{"summary":{"raw":"refactor: extract helper"}}]}`))
	}))

	out, err := runRepoCmd(t, "pr", "commits", "7")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"tighten validation", "extract helper"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrApprove(t *testing.T) {
	var gotMethod string
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		if !strings.HasSuffix(r.URL.Path, "/pullrequests/7/approve") {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{}`))
	}))

	out, err := runRepoCmd(t, "pr", "approve", "7")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if !strings.Contains(out, "Approved") {
		t.Errorf("output missing success message: %s", out)
	}
}

func TestPrApproveAll(t *testing.T) {
	approved := 0
	withTestServer(t, prRouter(map[string]http.HandlerFunc{
		"/repositories/acme/widgets/pullrequests": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"values":[{"id":3},{"id":5}]}`))
		},
		"/repositories/acme/widgets/pullrequests/3/approve": func(w http.ResponseWriter, r *http.Request) {
			approved++
			w.Write([]byte(`{}`))
		},
		"/repositories/acme/widgets/pullrequests/5/approve": func(w http.ResponseWriter, r *http.Request) {
			approved++
			w.Write([]byte(`{}`))
		},
	}))

	if _, err := runRepoCmd(t, "pr", "approve", "0"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if approved != 2 {
		t.Errorf("expected 2 PRs approved, got %d", approved)
	}
}

func FuzzParseBoolish(f *testing.F) {
	for _, seed := range []string{"true", "false", "1", "0", "", "TrUe", "no", "yes", "unresolved", "  ", "\x00\x01\x02"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, in string) {
		// Output type is bool — fuzzing just checks no panic occurs.
		_ = parseBoolish(in)
	})
}

func TestParseBoolish(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"true", true},
		{"1", true},
		{"yes", true},
		{"y", true},
		{"unresolved", true},
		{"TRUE", true},
		{"Yes", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"", false},
		{"nope", false},
	}
	for _, tt := range tests {
		if got := parseBoolish(tt.in); got != tt.want {
			t.Errorf("parseBoolish(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestPrRequestChanges(t *testing.T) {
	var gotMethod, gotPath string
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Write([]byte(`{}`))
	}))

	if _, err := runRepoCmd(t, "pr", "request-changes", "7"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/pullrequests/7/request-changes") {
		t.Errorf("path = %q", gotPath)
	}
}

func TestPrUnRequestChanges(t *testing.T) {
	var gotMethod, gotPath string
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Write([]byte(`{}`))
	}))

	if _, err := runRepoCmd(t, "pr", "no-request-changes", "7"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/pullrequests/7/request-changes") {
		t.Errorf("path = %q", gotPath)
	}
}

func TestPrUnApprove(t *testing.T) {
	var gotMethod string
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.Write([]byte(`{}`))
	}))

	if _, err := runRepoCmd(t, "pr", "no-approve", "7"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
}

func TestPrDecline(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"state":"DECLINED"}`))
	}))

	out, err := runRepoCmd(t, "pr", "decline", "7")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "OK") {
		t.Errorf("output missing OK: %s", out)
	}
}

func TestPrMerge(t *testing.T) {
	var body []byte
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Write([]byte(`{"state":"MERGED"}`))
	}))

	out, err := runRepoCmd(t, "pr", "merge", "7")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "MERGED") {
		t.Errorf("output missing MERGED: %s", out)
	}

	// Bitbucket answers a bodyless POST to this endpoint with a bare 400,
	// so the payload must never go out empty — see TestPrMergeFlags for
	// the field values.
	if len(body) == 0 {
		t.Fatal("merge request sent an empty body; Bitbucket rejects that with 400")
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("merge body is not valid JSON (%s): %v", body, err)
	}
	if got["merge_strategy"] != "merge_commit" {
		t.Errorf("merge_strategy = %v, want merge_commit", got["merge_strategy"])
	}
	if got["close_source_branch"] != false {
		t.Errorf("close_source_branch = %v, want false by default", got["close_source_branch"])
	}
}

func TestPrMergeFlags(t *testing.T) {
	var got map[string]any
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.Write([]byte(`{"state":"MERGED"}`))
	}))

	if _, err := runRepoCmd(t, "pr", "merge", "7", "--strategy", "squash", "--close-source-branch"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got["merge_strategy"] != "squash" {
		t.Errorf("merge_strategy = %v, want squash", got["merge_strategy"])
	}
	if got["close_source_branch"] != true {
		t.Errorf("close_source_branch = %v, want true", got["close_source_branch"])
	}
}

func TestPrView(t *testing.T) {
	var openedURL string
	saved := prViewOpener
	prViewOpener = func(url string) error {
		openedURL = url
		return nil
	}
	t.Cleanup(func() { prViewOpener = saved })

	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"links":{"html":{"href":"https://bitbucket.org/acme/widgets/pull-requests/7"}}}`))
	}))

	out, err := runRepoCmd(t, "pr", "view", "7")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if openedURL != "https://bitbucket.org/acme/widgets/pull-requests/7" {
		t.Errorf("opener got %q", openedURL)
	}
	if !strings.Contains(out, openedURL) {
		t.Errorf("output should print URL: %s", out)
	}
}

func TestPrCheckout(t *testing.T) {
	var gitArgs [][]string
	saved := prCheckoutGitExec
	prCheckoutGitExec = func(args ...string) ([]byte, error) {
		gitArgs = append(gitArgs, args)
		return nil, nil
	}
	t.Cleanup(func() { prCheckoutGitExec = saved })

	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"source":{"branch":{"name":"feature/foo"}}}`))
	}))

	out, err := runRepoCmd(t, "pr", "checkout", "7")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(gitArgs) != 2 {
		t.Fatalf("expected 2 git invocations (fetch + checkout), got %d: %v", len(gitArgs), gitArgs)
	}
	if gitArgs[0][0] != "fetch" || gitArgs[0][2] != "feature/foo" {
		t.Errorf("expected `git fetch origin feature/foo`, got %v", gitArgs[0])
	}
	if gitArgs[1][0] != "checkout" || gitArgs[1][1] != "feature/foo" {
		t.Errorf("expected `git checkout feature/foo`, got %v", gitArgs[1])
	}
	if !strings.Contains(out, "Checked out feature/foo") {
		t.Errorf("missing success message:\n%s", out)
	}
}

func TestPrComment(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Write([]byte(`{"id":42,"user":{"display_name":"Alice"}}`))
	}))

	out, err := runRepoCmd(t, "pr", "comment", "7", "looks good")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/pullrequests/7/comments") {
		t.Errorf("path = %q", gotPath)
	}
	content := gotBody["content"].(map[string]any)
	if content["raw"] != "looks good" {
		t.Errorf("content.raw = %v", content["raw"])
	}
	if !strings.Contains(out, "42") || !strings.Contains(out, "Alice") {
		t.Errorf("output missing id/author:\n%s", out)
	}
}

func TestPrCommentInline(t *testing.T) {
	var gotBody map[string]any
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Write([]byte(`{"id":99,"inline":{"path":"foo.go","to":42}}`))
	}))

	out, err := runRepoCmd(t, "pr", "comment-inline", "7", "foo.go", "42", "consider extracting this")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	inline := gotBody["inline"].(map[string]any)
	if inline["path"] != "foo.go" {
		t.Errorf("inline.path = %v, want foo.go", inline["path"])
	}
	if int(inline["to"].(float64)) != 42 {
		t.Errorf("inline.to = %v, want 42", inline["to"])
	}
	if !strings.Contains(out, "foo.go") || !strings.Contains(out, "42") {
		t.Errorf("output missing file/line:\n%s", out)
	}
}

func TestPrCommentInlineRejectsNonIntegerLine(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	}))

	_, err := runRepoCmd(t, "pr", "comment-inline", "7", "foo.go", "abc", "msg")
	if err == nil {
		t.Fatal("expected error on non-integer line")
	}
	if !strings.Contains(err.Error(), "line must be an integer") {
		t.Errorf("wrong error: %v", err)
	}
}

func TestPrCreate(t *testing.T) {
	var gotBody map[string]any
	var createdCount int
	withTestServer(t, prRouter(map[string]http.HandlerFunc{
		"/user": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"uuid":"{me}"}`))
		},
		"/repositories/acme/widgets/default-reviewers": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"values":[{"uuid":"{me}"},{"uuid":"{other}"}]}`))
		},
		"/repositories/acme/widgets/pullrequests": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				return
			}
			createdCount++
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &gotBody)
			w.Write([]byte(`{"id":42,"links":{"html":{"href":"https://bitbucket.org/acme/widgets/pull-requests/42"}}}`))
		},
	}))

	out, err := runRepoCmd(t, "pr", "create", "feature/foo", "main")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if createdCount != 1 {
		t.Errorf("expected 1 PR created, got %d", createdCount)
	}
	src := gotBody["source"].(map[string]any)["branch"].(map[string]any)
	if src["name"] != "feature/foo" {
		t.Errorf("source = %v", src)
	}
	// Default reviewers should exclude the current user.
	revs := gotBody["reviewers"].([]any)
	if len(revs) != 1 {
		t.Errorf("expected 1 reviewer (current user filtered), got %d: %+v", len(revs), revs)
	}
	if !strings.Contains(out, "42") || !strings.Contains(out, "https://bitbucket.org") {
		t.Errorf("output missing PR link/id:\n%s", out)
	}
}

func TestPrCreateMultipleDestinations(t *testing.T) {
	createCalls := 0
	withTestServer(t, prRouter(map[string]http.HandlerFunc{
		"/user": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"uuid":"{me}"}`))
		},
		"/repositories/acme/widgets/default-reviewers": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"values":[]}`))
		},
		"/repositories/acme/widgets/pullrequests": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				return
			}
			createCalls++
			w.Write([]byte(`{"id":1,"links":{"html":{"href":"x"}}}`))
		},
	}))

	if _, err := runRepoCmd(t, "pr", "create", "feature/foo", "main,develop"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if createCalls != 2 {
		t.Errorf("expected 2 PRs created (main + develop), got %d", createCalls)
	}
}

func TestPrCreateWithTitleAndDescription(t *testing.T) {
	var gotBody map[string]any
	withTestServer(t, prRouter(map[string]http.HandlerFunc{
		"/user": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"uuid":"{me}"}`))
		},
		"/repositories/acme/widgets/default-reviewers": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"values":[]}`))
		},
		"/repositories/acme/widgets/pullrequests": func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				b, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(b, &gotBody)
				w.Write([]byte(`{"id":1,"links":{"html":{"href":"x"}}}`))
			}
		},
	}))

	if _, err := runRepoCmd(t, "--title", "Hotfix!", "--description", "Fixes prod outage", "pr", "create", "feature/foo", "main"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotBody["title"] != "Hotfix!" {
		t.Errorf("title = %v, want Hotfix!", gotBody["title"])
	}
	if gotBody["description"] != "Fixes prod outage" {
		t.Errorf("description = %v", gotBody["description"])
	}
}

// chdirToTempGitRepo init's a fresh git repo on branch feature/local in
// a tempdir and chdirs into it. Returns once the repo is on disk so
// CurrentBranch() returns "feature/local".
func chdirToTempGitRepo(t *testing.T) {
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
	runGit("init", "-b", "feature/local", ".")
	if err := os.WriteFile("hello", []byte("hi"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	runGit("add", "hello")
	runGit("commit", "-m", "initial")
}

func TestPrCreateSingleArgUsesCurrentBranch(t *testing.T) {
	chdirToTempGitRepo(t)

	var gotBody map[string]any
	withTestServer(t, prRouter(map[string]http.HandlerFunc{
		"/user": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"uuid":"{me}"}`))
		},
		"/repositories/acme/widgets/default-reviewers": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"values":[]}`))
		},
		"/repositories/acme/widgets/pullrequests": func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				b, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(b, &gotBody)
				w.Write([]byte(`{"id":1,"links":{"html":{"href":"x"}}}`))
			}
		},
	}))

	// Single-arg form: only the destination is supplied; from-branch
	// must come from git HEAD.
	if _, err := runRepoCmd(t, "pr", "create", "main"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	src := gotBody["source"].(map[string]any)["branch"].(map[string]any)
	if src["name"] != "feature/local" {
		t.Errorf("expected from-branch=feature/local from git HEAD, got %v", src["name"])
	}
}

func TestPrCreateInteractivePromptsForTitle(t *testing.T) {
	saved := util.PromptFn
	util.PromptFn = func(q, def string) string {
		if strings.Contains(q, "title") {
			return "Interactive title"
		}
		return ""
	}
	t.Cleanup(func() { util.PromptFn = saved })

	var gotBody map[string]any
	withTestServer(t, prRouter(map[string]http.HandlerFunc{
		"/user": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"uuid":"{me}"}`))
		},
		"/repositories/acme/widgets/default-reviewers": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"values":[]}`))
		},
		"/repositories/acme/widgets/pullrequests": func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				b, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(b, &gotBody)
				w.Write([]byte(`{"id":1,"links":{"html":{"href":"x"}}}`))
			}
		},
	}))

	if _, err := runRepoCmd(t, "-i", "pr", "create", "feature/foo", "main"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotBody["title"] != "Interactive title" {
		t.Errorf("interactive title not used: %v", gotBody["title"])
	}
}

func TestPrCreateSkipDefaultReviewers(t *testing.T) {
	defaultReviewersHit := false
	withTestServer(t, prRouter(map[string]http.HandlerFunc{
		"/repositories/acme/widgets/default-reviewers": func(w http.ResponseWriter, r *http.Request) {
			defaultReviewersHit = true
			w.Write([]byte(`{"values":[]}`))
		},
		"/repositories/acme/widgets/pullrequests": func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"id":1,"links":{"html":{"href":"x"}}}`))
		},
	}))

	// Passing "0" for the add-default-reviewers parameter should skip
	// the default-reviewers fetch entirely.
	if _, err := runRepoCmd(t, "pr", "create", "feature/foo", "main", "0"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if defaultReviewersHit {
		t.Error("default-reviewers endpoint should not have been called")
	}
}
