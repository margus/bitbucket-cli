package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/margus/bb-cli/internal/config"
	"github.com/margus/bb-cli/internal/util"
)

func newTestClient(handler http.Handler) (*Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	c := NewWithAuth(&config.Auth{Username: "alice", AppPassword: "s3cret"}, srv.URL)
	return c, srv
}

func TestRequestSendsAuthHeader(t *testing.T) {
	var gotAuth, gotContentType, gotMethod, gotPath string
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	if _, err := c.Request("GET", "/user", nil, false); err != nil {
		t.Fatalf("Request error: %v", err)
	}

	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:s3cret"))
	if gotAuth != want {
		t.Errorf("auth header = %q, want %q", gotAuth, want)
	}
	if gotContentType != "application/json" {
		t.Errorf("content-type = %q, want application/json", gotContentType)
	}
	if gotMethod != "GET" {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/user" {
		t.Errorf("path = %q, want /user", gotPath)
	}
}

func TestRequestRepoURLPrefix(t *testing.T) {
	saved := util.ProjectURL
	defer func() { util.ProjectURL = saved }()
	util.ProjectURL = "acme/widgets"

	var gotPath string
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	if _, err := c.Request("GET", "/pullrequests", nil, true); err != nil {
		t.Fatalf("Request error: %v", err)
	}
	if want := "/repositories/acme/widgets/pullrequests"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestRequestPOSTBody(t *testing.T) {
	var gotBody []byte
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(201)
		w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()

	if _, err := c.Request("POST", "/x", map[string]string{"a": "b"}, false); err != nil {
		t.Fatalf("Request error: %v", err)
	}
	if !strings.Contains(string(gotBody), `"a":"b"`) {
		t.Errorf("body = %q, want it to contain {\"a\":\"b\"}", gotBody)
	}
}

func TestRequest401(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer srv.Close()

	_, err := c.Request("GET", "/x", nil, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "authorization") {
		t.Errorf("error = %v, want it to mention authorization", err)
	}
}

func TestRequest409IsPassthrough(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		w.Write([]byte(`{"conflict":true}`))
	}))
	defer srv.Close()

	b, err := c.Request("POST", "/x", nil, false)
	if err != nil {
		t.Fatalf("409 should not error, got: %v", err)
	}
	if !strings.Contains(string(b), "conflict") {
		t.Errorf("body = %q, want it to include the response", b)
	}
}

func TestRequestErrorEnvelope(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"type":"error","error":{"message":"branch missing"}}`))
	}))
	defer srv.Close()

	_, err := c.Request("POST", "/x", nil, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "branch missing") {
		t.Errorf("error = %v, want it to surface the API message", err)
	}
}

func TestRequestNon2xxNoEnvelope(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`oops`))
	}))
	defer srv.Close()

	_, err := c.Request("GET", "/x", nil, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %v, want it to mention status code", err)
	}
}

func TestJSONDecodesIntoStruct(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "alice", "id": 42})
	}))
	defer srv.Close()

	var out struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}
	if err := c.JSON("GET", "/x", nil, false, &out); err != nil {
		t.Fatalf("JSON error: %v", err)
	}
	if out.Name != "alice" || out.ID != 42 {
		t.Errorf("decoded = %+v, want {Name:alice ID:42}", out)
	}
}

func TestJSONNonJSONIntoStringPointer(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("diff --git a b\nplain text"))
	}))
	defer srv.Close()

	var s string
	if err := c.JSON("GET", "/diff", nil, false, &s); err != nil {
		t.Fatalf("JSON error: %v", err)
	}
	if !strings.Contains(s, "diff --git") {
		t.Errorf("decoded string = %q, want raw body", s)
	}
}

func TestNewWithAuthDefaultsBaseURL(t *testing.T) {
	c := NewWithAuth(&config.Auth{Username: "a", AppPassword: "p"}, "")
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
}
