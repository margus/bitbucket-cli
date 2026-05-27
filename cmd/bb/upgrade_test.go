package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
)

func runtimeGOOS() string   { return runtime.GOOS }
func runtimeGOARCH() string { return runtime.GOARCH }

func TestUpgradeAlreadyLatest(t *testing.T) {
	// Pin the build-version to something lexically > any release tag we
	// return below, so the "already latest" short-circuit fires.
	savedVer := version
	version = "v9.9.9"
	t.Cleanup(func() { version = savedVer })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"tag_name":"v0.0.0","assets":[]}`))
	}))
	defer srv.Close()

	saved := upgradeReleasesURL
	upgradeReleasesURL = srv.URL
	t.Cleanup(func() { upgradeReleasesURL = saved })

	out, err := runCmd(t, "upgrade")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "already on the latest") {
		t.Errorf("output missing 'already on latest':\n%s", out)
	}
}

func TestUpgradeNoAssetForOSArch(t *testing.T) {
	// Newer version exists but no asset matches our runtime.GOOS/GOARCH.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"tag_name":"v9.9.9","assets":[{"name":"bb-haiku-amd64","browser_download_url":"https://example.com/bb"}]}`))
	}))
	defer srv.Close()

	saved := upgradeReleasesURL
	upgradeReleasesURL = srv.URL
	t.Cleanup(func() { upgradeReleasesURL = saved })

	_, err := runCmd(t, "upgrade")
	if err == nil {
		t.Fatal("expected error when no matching asset exists")
	}
	if !strings.Contains(err.Error(), "no release asset") {
		t.Errorf("error = %v, want it to mention no asset", err)
	}
}

func TestUpgradeDownloadsBinary(t *testing.T) {
	// Two mock servers: one returns the release manifest, one serves the
	// binary payload.
	binSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("#!/bin/sh\necho v9.9.9\n"))
	}))
	defer binSrv.Close()

	assetName := "bb-" + runtimeGOOS() + "-" + runtimeGOARCH()
	releaseBody := `{
      "tag_name":"v9.9.9",
      "assets":[{"name":"` + assetName + `","browser_download_url":"` + binSrv.URL + `/bin"}]
    }`
	relSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(releaseBody))
	}))
	defer relSrv.Close()

	saved := upgradeReleasesURL
	upgradeReleasesURL = relSrv.URL
	t.Cleanup(func() { upgradeReleasesURL = saved })

	// Use a tempfile as the "install path".
	tmp, err := os.CreateTemp(t.TempDir(), "bb-target-*")
	if err != nil {
		t.Fatalf("temp: %v", err)
	}
	tmp.Close()
	savedTarget := upgradeTargetPath
	upgradeTargetPath = func() (string, error) { return tmp.Name(), nil }
	t.Cleanup(func() { upgradeTargetPath = savedTarget })

	// And pin current version below v9.9.9.
	savedVer := version
	version = "v0.0.0"
	t.Cleanup(func() { version = savedVer })

	out, err := runCmd(t, "upgrade")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "updated") {
		t.Errorf("output missing 'updated' marker:\n%s", out)
	}
	// The "binary" we wrote contains a known marker.
	content, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatalf("read installed file: %v", err)
	}
	if !strings.Contains(string(content), "v9.9.9") {
		t.Errorf("installed file doesn't look like the downloaded binary:\n%s", content)
	}
}

func TestUpgradeFetchReleaseFails(t *testing.T) {
	// Server that 500s on every request.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "kaboom", 500)
	}))
	defer srv.Close()

	saved := upgradeReleasesURL
	upgradeReleasesURL = srv.URL
	t.Cleanup(func() { upgradeReleasesURL = saved })

	// 500 doesn't fail the http.Get itself, but the body is "kaboom" which
	// isn't JSON — should surface as a parse error.
	_, err := runCmd(t, "upgrade")
	if err == nil {
		t.Fatal("expected upgrade to surface release-fetch failure")
	}
}

func TestUpgradeBadReleaseJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	saved := upgradeReleasesURL
	upgradeReleasesURL = srv.URL
	t.Cleanup(func() { upgradeReleasesURL = saved })

	_, err := runCmd(t, "upgrade")
	if err == nil {
		t.Fatal("expected error on bad JSON")
	}
	if !strings.Contains(err.Error(), "parsing release") {
		t.Errorf("error = %v, want 'parsing release'", err)
	}
}
