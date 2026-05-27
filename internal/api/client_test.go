package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/margus/bitbucket-cli/internal/config"
	"github.com/margus/bitbucket-cli/internal/util"
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

// When an access token is set, it should be used as a Bearer token
// instead of Basic auth — even if username/password are also present.
func TestRequestPrefersBearerOverBasic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer abc123" {
			t.Errorf("auth header = %q, want %q", got, "Bearer abc123")
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewWithAuth(&config.Auth{
		Username:    "alice", // present but should be ignored
		AppPassword: "s3cret",
		AccessToken: "abc123",
	}, srv.URL)
	if _, err := c.Request("GET", "/user", nil, false); err != nil {
		t.Fatalf("Request error: %v", err)
	}
}

// With only an access token set, Bearer is used.
func TestRequestBearerOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer t0k3n" {
			t.Errorf("auth header = %q, want %q", got, "Bearer t0k3n")
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewWithAuth(&config.Auth{AccessToken: "t0k3n"}, srv.URL)
	if _, err := c.Request("GET", "/user", nil, false); err != nil {
		t.Fatalf("Request error: %v", err)
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

func TestRetryOn5xx(t *testing.T) {
	saved := MaxRetries
	savedDelay := RetryBaseDelay
	MaxRetries = 2
	RetryBaseDelay = 0
	t.Cleanup(func() { MaxRetries = saved; RetryBaseDelay = savedDelay })

	var attempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			http.Error(w, "transient", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewWithAuth(&config.Auth{AccessToken: "t"}, srv.URL)
	if _, err := c.Request("GET", "/x", nil, false); err != nil {
		t.Fatalf("retry should have eventually succeeded: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts (1 initial + 2 retries), got %d", attempts)
	}
}

func TestRetryExhausts(t *testing.T) {
	saved := MaxRetries
	savedDelay := RetryBaseDelay
	MaxRetries = 1
	RetryBaseDelay = 0
	t.Cleanup(func() { MaxRetries = saved; RetryBaseDelay = savedDelay })

	var attempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		http.Error(w, "always broken", 500)
	}))
	defer srv.Close()

	c := NewWithAuth(&config.Auth{AccessToken: "t"}, srv.URL)
	_, err := c.Request("GET", "/x", nil, false)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts (1 + 1 retry), got %d", attempts)
	}
}

func TestDebugHook(t *testing.T) {
	var calls []string
	saved := Debug
	Debug = func(method, url string, status int, body []byte) {
		calls = append(calls, fmt.Sprintf("%s %d", method, status))
	}
	t.Cleanup(func() { Debug = saved })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewWithAuth(&config.Auth{AccessToken: "t"}, srv.URL)
	if _, err := c.Request("GET", "/x", nil, false); err != nil {
		t.Fatalf("request: %v", err)
	}
	if len(calls) != 1 || calls[0] != "GET 200" {
		t.Errorf("debug hook saw %v, want one call %q", calls, "GET 200")
	}
}

func TestNewFromConfigReturnsErrNoAuthWhenMissing(t *testing.T) {
	// Sandbox HOME so config.Init() finds an empty config.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("BB_AUTH_USERNAME", "")
	t.Setenv("BB_AUTH_APPPASSWORD", "")
	t.Setenv("BB_AUTH_ACCESSTOKEN", "")
	config.Reset()
	t.Cleanup(config.Reset)

	c, err := NewFromConfig()
	if c != nil {
		t.Errorf("expected nil client, got %v", c)
	}
	if err == nil {
		t.Fatal("expected ErrNoAuth, got nil")
	}
	if !errors.Is(err, ErrNoAuth) {
		t.Errorf("expected ErrNoAuth, got %v", err)
	}
}

func TestJSONReturnsErrorOnMalformedBody(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not valid json`))
	}))
	defer srv.Close()

	var out struct {
		Foo string `json:"foo"`
	}
	err := c.JSON("GET", "/x", nil, false, &out)
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
	if !strings.Contains(err.Error(), "decoding response") {
		t.Errorf("error = %v, want it to mention 'decoding response'", err)
	}
}

func TestJSONEmptyBodyIsOK(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bitbucket's POST .../approve returns no body sometimes.
		w.WriteHeader(204)
	}))
	defer srv.Close()

	var out struct {
		Foo string `json:"foo"`
	}
	if err := c.JSON("POST", "/x", nil, false, &out); err != nil {
		t.Fatalf("empty body should not error: %v", err)
	}
}

func TestRequestNilOutDiscardsBody(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"any":"thing"}`))
	}))
	defer srv.Close()

	// out=nil means "I just want to know it succeeded".
	if err := c.JSON("GET", "/x", nil, false, nil); err != nil {
		t.Fatalf("JSON(out=nil) = %v", err)
	}
}

func TestNewFromConfigPicksUpEnvAuth(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("BB_AUTH_ACCESSTOKEN", "ATBB-fromenv")
	config.Reset()
	t.Cleanup(config.Reset)

	c, err := NewFromConfig()
	if err != nil {
		t.Fatalf("NewFromConfig: %v", err)
	}
	if c.auth == nil || c.auth.AccessToken != "ATBB-fromenv" {
		t.Errorf("auth not picked up from env: %+v", c.auth)
	}
}
