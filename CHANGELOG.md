# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project tracks [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
once it leaves 0.x.

## [Unreleased]

## [0.1.0] - 2026-08-21

This is the initial public release of `bitbucket-cli` — a Bitbucket Cloud REST API CLI written in Go.

### Fixed

- `bb pr merge` always failed with a bare `400 Bad Request`. The request went out with
  `Content-Type: application/json` but no body at all, which Bitbucket rejects; it now sends a
  proper merge payload. This was the only bodyless POST in the CLI, so no other command was
  affected.

### Added

- `bb pr merge --strategy {merge_commit,squash,fast_forward}` (default `merge_commit`) and
  `--close-source-branch` (default off — the source branch is kept unless you ask for it).

### Commands

- `bb auth {token,save,logout,show}` — manage credentials (access token or app password)
- `bb branch {list,user,name}` — list and filter branches
- `bb browse {browse,url}` — open the repo in the browser or print its URL
- `bb env {list,variables,create-variable,update-variable}` — manage deployment environments and variables
- `bb pipeline {get,latest,wait,run,custom,logs}` — run, watch, and inspect pipelines
- `bb pr {list,show,diff,files,commits,approve,no-approve,request-changes,no-request-changes,decline,merge,create,comment,comment-inline,checkout,view}` — full PR lifecycle
- `bb repo list [workspace]` — browse repositories in a workspace
- `bb workspace list` — list workspaces the authenticated user has access to
- `bb completion {bash,zsh,fish,powershell}` — shell completion scripts
- `bb upgrade` — self-update from the GitHub Releases page

### Global flags

- `--project owner/repo` — work against any repo without a checkout
- `--output / -o {json,yaml}` — machine-readable output for `list` / `get` commands
- `--debug` — log every HTTP exchange to stderr
- `-i, --interactive`, `--title`, `--description` — non-default PR-creation flow

### Auth

- Bitbucket workspace / repository / project **access tokens** (Bearer auth) — the recommended path
- Legacy **username + app password** (HTTP Basic) — still supported for older setups
- Env-var overrides (`BB_AUTH_ACCESSTOKEN`, `BB_AUTH_USERNAME`, `BB_AUTH_APPPASSWORD`) for CI use
- Config at `$XDG_CONFIG_HOME/bb/config.json` (default `~/.config/bb/config.json`), `0600` permissions

### API endpoints

Every endpoint the CLI hits has been audited against Atlassian's CHANGE-2770 deprecation (removal of cross-workspace REST APIs, effective 2026-04-14):

- `bb workspace list` calls the new `GET /2.0/user/workspaces` (released January 2026)
- All other commands use workspace-scoped paths (`/repositories/{ws}/{repo}/...`) that were unaffected
- `bb repo list <workspace>` uses the per-workspace listing, not the deprecated cross-workspace one

### Internals

- Built on [cobra](https://github.com/spf13/cobra) (command dispatch) + [viper](https://github.com/spf13/viper) (config + env-var overrides)
- HTTP retries on 5xx with exponential backoff (configurable via `api.MaxRetries`)
- Pure stdlib HTTP client; single static binary

### Distribution

- Pre-built binaries for linux/darwin/windows × amd64/arm64
- Homebrew tap (`brew install margus/tap/bb`)
- Docker image (`ghcr.io/margus/bitbucket-cli:latest`)
- `go install github.com/margus/bitbucket-cli/cmd/bb@latest`
- Self-update via `bb upgrade`

### Supply chain

- Releases include CycloneDX SBOMs and sigstore/cosign keyless signatures
- CI runs `govulncheck`, `golangci-lint` (errcheck, gosec, govet, ineffassign, misspell, staticcheck, unused), and race-detector tests on linux/macos/windows
- Dependabot watches `github-actions` and `gomod` weekly

### Documentation

- README is install + quickstart + links
- Full docs under `docs/` — getting started, auth, configuration, shell completion, per-command reference, development, release process
- `CHANGELOG.md`, `CONTRIBUTING.md`, `SECURITY.md`

[Unreleased]: https://github.com/margus/bitbucket-cli/compare/v0.1.0...main
[0.1.0]: https://github.com/margus/bitbucket-cli/releases/tag/v0.1.0
