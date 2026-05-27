package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

const pipelineCompletedBody = `{
  "creator":{"display_name":"Alice"},
  "repository":{"name":"widgets"},
  "target":{"ref_name":"main"},
  "state":{"name":"COMPLETED","result":{"name":"SUCCESSFUL"}},
  "created_on":"2026-05-26T10:00:00Z",
  "completed_on":"2026-05-26T10:05:00Z",
  "build_number":42
}`

func TestPipelineGet(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/pipelines/42") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(pipelineCompletedBody))
	}))

	out, err := runRepoCmd(t, "pipeline", "get", "42")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"COMPLETED", "SUCCESSFUL", "Alice", "main"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPipelineLatest(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/pipelines/") {
			w.Write([]byte(`{"size":7}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/pipelines/7") {
			w.Write([]byte(pipelineCompletedBody))
			return
		}
		t.Errorf("unexpected path: %s", r.URL.Path)
	}))

	out, err := runRepoCmd(t, "pipeline", "latest")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "COMPLETED") {
		t.Errorf("output missing COMPLETED:\n%s", out)
	}
}

func TestPipelineWaitPollsUntilComplete(t *testing.T) {
	saved := pipelineWaitSleep
	pipelineWaitSleep = 0
	t.Cleanup(func() { pipelineWaitSleep = saved })

	var hits int
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.Write([]byte(`{"state":{"name":"RUNNING","result":{"name":""}}}`))
			return
		}
		w.Write([]byte(pipelineCompletedBody))
	}))

	out, err := runRepoCmd(t, "pipeline", "wait", "42")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if hits < 3 {
		t.Errorf("expected poll loop to call backend at least 3 times, got %d", hits)
	}
	if !strings.Contains(out, "COMPLETED") {
		t.Errorf("final output missing COMPLETED:\n%s", out)
	}
}

func TestPipelineWaitWithoutNumberPicksLatest(t *testing.T) {
	saved := pipelineWaitSleep
	pipelineWaitSleep = 0
	t.Cleanup(func() { pipelineWaitSleep = saved })

	var hits int
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		// First call: size endpoint to discover latest pipeline.
		if strings.HasSuffix(r.URL.Path, "/pipelines/") {
			w.Write([]byte(`{"size":99}`))
			return
		}
		// Subsequent calls hit /pipelines/99 — return COMPLETED right
		// away so the wait loop exits immediately.
		w.Write([]byte(pipelineCompletedBody))
	}))

	out, err := runRepoCmd(t, "pipeline", "wait")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "Pipeline: 99") {
		t.Errorf("expected derived pipeline id in output:\n%s", out)
	}
	if hits < 2 {
		t.Errorf("expected at least 2 backend calls, got %d", hits)
	}
}

func TestPipelineRun(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Write([]byte(pipelineCompletedBody))
	}))

	if _, err := runRepoCmd(t, "pipeline", "run", "main"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/pipelines/") {
		t.Errorf("path = %q", gotPath)
	}
	target, _ := gotBody["target"].(map[string]any)
	if target["ref_name"] != "main" || target["ref_type"] != "branch" {
		t.Errorf("bad payload target: %+v", target)
	}
}

func TestPipelineLogsSingleStep(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/pipelines/42/steps/{step-uuid}/log") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte("running tests...\nall green\n"))
	}))

	out, err := runRepoCmd(t, "pipeline", "logs", "42", "{step-uuid}")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"running tests", "all green"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPipelineLogsAllSteps(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Step listing.
		if strings.HasSuffix(r.URL.Path, "/pipelines/42/steps/") {
			w.Write([]byte(`{"values":[{"uuid":"{s1}","name":"build"},{"uuid":"{s2}","name":"test"}]}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/{s1}/log") {
			w.Write([]byte("compiled\n"))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/{s2}/log") {
			w.Write([]byte("tested\n"))
			return
		}
		t.Errorf("unexpected path: %s", r.URL.Path)
	}))

	out, err := runRepoCmd(t, "pipeline", "logs", "42")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"step: build", "compiled", "step: test", "tested"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPipelineCustom(t *testing.T) {
	var gotBody map[string]any
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Write([]byte(`{"build_number":99}`))
	}))

	out, err := runRepoCmd(t, "pipeline", "custom", "main", "deploy")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	target := gotBody["target"].(map[string]any)
	sel := target["selector"].(map[string]any)
	if sel["type"] != "custom" || sel["pattern"] != "deploy" {
		t.Errorf("bad selector: %+v", sel)
	}
	if !strings.Contains(out, "/results/99") {
		t.Errorf("output missing build number link:\n%s", out)
	}
}
