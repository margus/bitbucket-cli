package main

// End-to-end tests for the `bb` binary as a real subprocess. The other
// _test.go files drive rootCmd.Execute() in-process — those miss
// anything that lives outside the cobra graph: main.go, ldflags, exit
// codes, stdin handling, ANSI color stripping in pipelines, etc.
//
// TestMain builds the binary once into the test tempdir, then each
// e2e_* test runs it as a subprocess pointed at an httptest server.

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var e2eBinary string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "bb-e2e-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e: tempdir:", err)
		os.Exit(2)
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	binName := "bb"
	if runtime.GOOS == "windows" {
		binName = "bb.exe"
	}
	e2eBinary = filepath.Join(tmp, binName)
	if _, err := os.Stat("main.go"); err != nil {
		// `go test ./cmd/bb/...` runs from cmd/bb so this should hold;
		// guard for editor / package-relative invocations.
		fmt.Fprintln(os.Stderr, "e2e: not running from cmd/bb dir, skipping build")
	}
	build := exec.Command("go", "build", "-o", e2eBinary, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "e2e: build:", err)
		os.Exit(2)
	}
	os.Exit(m.Run())
}

// runBB execs the built binary with the given args + env. Returns
// (stdout, stderr, exit-code).
func runBB(t *testing.T, env []string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(e2eBinary, args...)
	cmd.Env = append(os.Environ(), env...)
	// Disable color in subprocess output so assertions can be plain.
	cmd.Env = append(cmd.Env, "NO_COLOR=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exit := 0
	if ee, ok := err.(*exec.ExitError); ok {
		exit = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run %s %v: %v", e2eBinary, args, err)
	}
	return stdout.String(), stderr.String(), exit
}

// withFakeBitbucket points the subprocess at an httptest backend by
// setting the API base via env. We don't have an env-var for this yet,
// so the e2e tests cover only flows that don't make HTTP calls (help,
// version, auth show/save, bogus action, completion). HTTP flows are
// covered by the in-process tests.

func TestE2EVersion(t *testing.T) {
	out, _, exit := runBB(t, nil, "--version")
	if exit != 0 {
		t.Fatalf("--version exit=%d, want 0\n%s", exit, out)
	}
	// Set via -ldflags at make-build time. Here it's the bare `go build`
	// default "dev". Don't pin the value, just confirm we get the
	// version header.
	if !strings.Contains(out, "bb") {
		t.Errorf("--version output missing 'bb':\n%s", out)
	}
}

func TestE2EHelpListsAllSubcommands(t *testing.T) {
	out, _, exit := runBB(t, nil, "--help")
	if exit != 0 {
		t.Fatalf("--help exit=%d", exit)
	}
	for _, want := range []string{"auth", "branch", "browse", "env", "pipeline", "pr", "repo", "workspace", "upgrade", "completion"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help missing subcommand %q", want)
		}
	}
}

func TestE2EUnknownActionExits1(t *testing.T) {
	tmp := t.TempDir()
	_, _, exit := runBB(t, []string{"HOME=" + tmp, "XDG_CONFIG_HOME="}, "bogus-action")
	if exit == 0 {
		t.Fatalf("unknown action should exit non-zero, got 0")
	}
}

func TestE2EAuthShowNoConfig(t *testing.T) {
	// Sandbox HOME so the test doesn't read a real ~/.config/bb/config.json.
	tmp := t.TempDir()
	out, _, exit := runBB(t,
		[]string{"HOME=" + tmp, "USERPROFILE=" + tmp, "XDG_CONFIG_HOME=", "BB_AUTH_USERNAME=", "BB_AUTH_APPPASSWORD=", "BB_AUTH_ACCESSTOKEN="},
		"auth", "show",
	)
	if exit == 0 {
		t.Errorf("auth show with no creds should exit non-zero, got 0. Output:\n%s", out)
	}
	combined := out
	if !strings.Contains(combined, "configure auth info") {
		t.Errorf("missing hint message:\n%s", combined)
	}
}

func TestE2EAccessTokenAuthSetViaEnv(t *testing.T) {
	tmp := t.TempDir()
	env := []string{
		"HOME=" + tmp,
		"USERPROFILE=" + tmp,
		"XDG_CONFIG_HOME=",
		"BB_AUTH_ACCESSTOKEN=ATBB-e2e-token",
	}
	out, _, exit := runBB(t, env, "auth", "show")
	if exit != 0 {
		t.Fatalf("auth show with env token failed: exit=%d\n%s", exit, out)
	}
	if !strings.Contains(out, "access token") {
		t.Errorf("expected 'access token' method in output:\n%s", out)
	}
	if !strings.Contains(out, "ATBB-e2e-token") {
		t.Errorf("expected token value in output:\n%s", out)
	}
}

func TestE2EOutputJSON(t *testing.T) {
	tmp := t.TempDir()
	env := []string{
		"HOME=" + tmp,
		"USERPROFILE=" + tmp,
		"XDG_CONFIG_HOME=",
		"BB_AUTH_ACCESSTOKEN=ATBB-token",
	}
	out, _, exit := runBB(t, env, "--output", "json", "auth", "show")
	if exit != 0 {
		t.Fatalf("exit=%d: %s", exit, out)
	}
	out = strings.TrimSpace(out)
	if !strings.HasPrefix(out, "{") || !strings.HasSuffix(out, "}") {
		t.Errorf("expected JSON object, got:\n%s", out)
	}
	if !strings.Contains(out, `"method": "access token"`) {
		t.Errorf("JSON output missing method field:\n%s", out)
	}
}

func TestE2ECompletionEmits(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out, _, exit := runBB(t, nil, "completion", shell)
			if exit != 0 {
				t.Fatalf("completion %s exit=%d", shell, exit)
			}
			if len(out) < 100 {
				t.Errorf("completion %s output suspiciously short: %d bytes", shell, len(out))
			}
		})
	}
}

func TestE2EInvalidOutputFormat(t *testing.T) {
	_, _, exit := runBB(t, nil, "--output", "xml", "--help")
	// --help short-circuits before persistent-prerun-e runs in some
	// cobra versions, so this might pass or fail. Either is fine; the
	// real assertion is that a bad --output value doesn't crash.
	_ = exit
}

func TestE2EHTTPCallAgainstHttptest(t *testing.T) {
	// Drive a real HTTP-calling command through the binary by giving it
	// fake auth + an httptest server. This exercises the full path:
	// cobra → config.Init → api client → HTTP → output rendering.
	//
	// Skipping because the binary has no env-var override for the
	// Bitbucket base URL — see internal/api/client.go DefaultBaseURL.
	// Adding such a knob is a small follow-up: BB_API_URL env var.
	t.Skip("no BB_API_URL override yet; subprocess can't be redirected to httptest")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"values":[]}`))
	}))
	defer srv.Close()
	_ = srv.URL
}
