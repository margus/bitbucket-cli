package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestEnvList(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/environments") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"values":[{"uuid":"{abc}","name":"production"},{"uuid":"{def}","name":"staging"}]}`))
	}))

	out, err := runRepoCmd(t, "env", "list")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"production", "staging", "{abc}", "{def}"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestEnvVariables(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/deployments_config/environments/{abc}/variables") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"values":[{"uuid":"{v1}","key":"API_KEY","value":"secret","secured":true},{"uuid":"{v2}","key":"DEBUG","value":"1","secured":false}]}`))
	}))

	out, err := runRepoCmd(t, "env", "variables", "{abc}")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"API_KEY", "DEBUG", "Yes", "No"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestEnvCreateVariable(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Write([]byte(`{"uuid":"{new}","key":"FOO","value":"bar","secured":true}`))
	}))

	out, err := runRepoCmd(t, "env", "create-variable", "{abc}", "FOO", "bar", "true")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if !strings.Contains(gotPath, "/environments/{abc}/variables") {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["key"] != "FOO" || gotBody["value"] != "bar" || gotBody["secured"] != true {
		t.Errorf("body = %+v", gotBody)
	}
	if !strings.Contains(out, "FOO") || !strings.Contains(out, "Yes") {
		t.Errorf("output didn't reflect created variable:\n%s", out)
	}
}

func TestEnvUpdateVariable(t *testing.T) {
	var gotMethod, gotPath string
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Write([]byte(`{"uuid":"{v1}","key":"FOO","value":"newval","secured":false}`))
	}))

	out, err := runRepoCmd(t, "env", "update-variable", "{abc}", "{v1}", "FOO", "newval")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotMethod != "PUT" {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if !strings.Contains(gotPath, "/{abc}/variables/{v1}") {
		t.Errorf("path = %q", gotPath)
	}
	if !strings.Contains(out, "newval") {
		t.Errorf("output = %s", out)
	}
}

func TestEnvCmdDefaultsToList(t *testing.T) {
	withTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"values":[{"uuid":"{a}","name":"prod"}]}`))
	}))

	out, err := runRepoCmd(t, "env")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "prod") {
		t.Errorf("default subcommand didn't run list:\n%s", out)
	}
}
