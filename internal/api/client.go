package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/margus/bb-cli/internal/config"
	"github.com/margus/bb-cli/internal/util"
)

const baseURL = "https://api.bitbucket.org/2.0"

// Client makes authenticated Bitbucket REST requests.
type Client struct {
	httpc *http.Client
	auth  *config.Auth
}

// New returns a client backed by the user's saved auth; if no auth is
// set it prints a hint and exits.
func New() *Client {
	cfg, err := config.Load()
	if err != nil {
		util.Errorln("%v", err)
		os.Exit(1)
	}
	if cfg.Auth == nil || cfg.Auth.Username == "" {
		util.O("You have to configure auth info to use this command.", "red")
		util.O(`Run "bb auth" first.`, "yellow")
		os.Exit(1)
	}
	return &Client{
		httpc: &http.Client{Timeout: 60 * time.Second},
		auth:  cfg.Auth,
	}
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

	var body io.Reader
	if method != http.MethodGet && payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encoding payload: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, baseURL+urlPath, body)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	authStr := c.auth.Username + ":" + c.auth.AppPassword
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(authStr)))

	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("authorization error, please check your credentials")
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if resp.StatusCode == 409 {
			return respBody, nil
		}
		// Try to surface API error message.
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
		return nil, fmt.Errorf("an error occurred, status code: %d", resp.StatusCode)
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
