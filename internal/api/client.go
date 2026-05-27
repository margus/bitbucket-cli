package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/margus/bitbucket-cli/internal/config"
	"github.com/margus/bitbucket-cli/internal/util"
)

// DefaultBaseURL is the production Bitbucket Cloud REST API root. Tests
// can construct a Client with a different base via NewWithAuth.
const DefaultBaseURL = "https://api.bitbucket.org/2.0"

// Debug, if non-nil, is invoked once per HTTP exchange with a short
// summary (method, status, latency, URL). Wire it from cmd/bb when
// --debug is set.
var Debug func(method, url string, status int, body []byte)

// MaxRetries is the number of times a 5xx response is retried before
// surfacing an error. Each retry waits RetryBaseDelay * 2^attempt.
var MaxRetries = 2
var RetryBaseDelay = 500 * time.Millisecond

// Client makes authenticated Bitbucket REST requests.
type Client struct {
	httpc   *http.Client
	auth    *config.Auth
	baseURL string
}

// NewWithAuth builds a Client with explicit auth and base URL. baseURL
// may be empty, in which case DefaultBaseURL is used.
func NewWithAuth(auth *config.Auth, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		httpc:   &http.Client{Timeout: 60 * time.Second},
		auth:    auth,
		baseURL: baseURL,
	}
}

// ErrNoAuth is returned by NewFromConfig when the config file and env
// vars don't yield a usable credential. Callers that want to print a
// friendly message and exit should check for this with errors.Is.
var ErrNoAuth = errors.New("no Bitbucket auth configured")

// NewFromConfig builds a Client from viper-loaded config. Returns
// ErrNoAuth when neither an access token nor a username is set, and
// any config-read error verbatim.
func NewFromConfig() (*Client, error) {
	if err := config.Init(); err != nil {
		return nil, err
	}
	auth := config.LoadAuth()
	if auth == nil {
		return nil, ErrNoAuth
	}
	return NewWithAuth(auth, ""), nil
}

// Request is the low-level call. If isRepoURL is true, urlPath is prefixed
// with /repositories/<owner>/<repo>. Returns the raw response body; non-2xx
// statuses (except 409) become errors.
func (c *Client) Request(method, urlPath string, payload any, isRepoURL bool) ([]byte, error) {
	if isRepoURL {
		repo, err := util.RepoPath()
		if err != nil {
			return nil, err
		}
		urlPath = "/repositories/" + repo + urlPath
	}

	var bodyBytes []byte
	if method != http.MethodGet && payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encoding payload: %w", err)
		}
		bodyBytes = b
	}

	var respBody []byte
	var status int
	url := c.baseURL + urlPath

	for attempt := 0; ; attempt++ {
		var body io.Reader
		if bodyBytes != nil {
			body = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, fmt.Errorf("building request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if c.auth.AccessToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.auth.AccessToken)
		} else {
			authStr := c.auth.Username + ":" + c.auth.AppPassword
			req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(authStr)))
		}

		resp, err := c.httpc.Do(req)
		if err != nil {
			if attempt < MaxRetries {
				time.Sleep(RetryBaseDelay << attempt)
				continue
			}
			return nil, fmt.Errorf("request: %w", err)
		}
		respBody, err = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		status = resp.StatusCode
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}
		if Debug != nil {
			Debug(method, url, status, respBody)
		}
		// Retry 5xx (server-side, likely transient).
		if status >= 500 && status < 600 && attempt < MaxRetries {
			time.Sleep(RetryBaseDelay << attempt)
			continue
		}
		break
	}

	if status == 401 {
		return nil, fmt.Errorf("authorization error (401): the saved credentials were rejected. Run `bb auth show` to check what's stored, or `bb auth token` to refresh")
	}
	if status == 403 {
		return nil, fmt.Errorf("forbidden (403): the credentials are valid but lack the required scope. Re-create the token with `pullrequest`, `pipeline`, `repository`, and/or `workspace` scopes as appropriate")
	}
	if status < 200 || status > 299 {
		if status == 409 {
			return respBody, nil
		}
		var errEnvelope struct {
			Type  string `json:"type"`
			Error struct {
				Message string `json:"message"`
				Detail  string `json:"detail"`
			} `json:"error"`
		}
		if json.Unmarshal(respBody, &errEnvelope) == nil && errEnvelope.Error.Message != "" {
			return nil, fmt.Errorf("%s", errEnvelope.Error.Message)
		}
		fmt.Println(string(respBody))
		return nil, fmt.Errorf("an error occurred, status code: %d", status)
	}

	return respBody, nil
}

// JSON sends a request and decodes the JSON body into out.
func (c *Client) JSON(method, urlPath string, payload any, isRepoURL bool, out any) error {
	b, err := c.Request(method, urlPath, payload, isRepoURL)
	if err != nil {
		return err
	}
	if len(b) == 0 || out == nil {
		return nil
	}
	if err := json.Unmarshal(b, out); err != nil {
		// Non-JSON body: callers that pass *string get the raw text.
		if sp, ok := out.(*string); ok {
			*sp = string(b)
			return nil
		}
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}
