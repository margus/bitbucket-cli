# Getting started

A 30-second walkthrough from "I just downloaded bb" to "I can ship PRs".

## 1. Install

Pre-built binary (recommended):

```sh
# macOS arm64
curl -L -o /usr/local/bin/bb https://github.com/margus/bitbucket-cli/releases/latest/download/bb-darwin-arm64
chmod +x /usr/local/bin/bb

# macOS amd64
curl -L -o /usr/local/bin/bb https://github.com/margus/bitbucket-cli/releases/latest/download/bb-darwin-amd64
chmod +x /usr/local/bin/bb

# Linux amd64
curl -L -o /usr/local/bin/bb https://github.com/margus/bitbucket-cli/releases/latest/download/bb-linux-amd64
chmod +x /usr/local/bin/bb

# Linux arm64
curl -L -o /usr/local/bin/bb https://github.com/margus/bitbucket-cli/releases/latest/download/bb-linux-arm64
chmod +x /usr/local/bin/bb
```

Windows: download `bb-windows-amd64.exe` from
<https://github.com/margus/bitbucket-cli/releases/latest> and put it on
your PATH.

Verify:

```sh
bb version
```

Other options:

```sh
# Go modules
go install github.com/margus/bitbucket-cli/cmd/bb@latest

# Homebrew (macOS / Linux)
brew install margus/tap/bb

# Docker (great for CI)
docker run --rm ghcr.io/margus/bitbucket-cli:latest version

# From source
git clone https://github.com/margus/bitbucket-cli && cd bitbucket-cli
make build
```

## 2. Authenticate

The recommended path is a Bitbucket **workspace access token** (Atlassian
docs: <https://support.atlassian.com/bitbucket-cloud/docs/access-tokens/>).

In the Bitbucket UI: *Workspace settings → Access tokens → Create token*.

Give it the scopes you need (typically `pullrequest:read+write`,
`pipeline:read+write`, `repository:read`), then:

```sh
bb auth token   # paste the token when prompted
```

If you prefer the legacy username + app password flow, see
[`auth.md`](auth.md).

Verify:

```sh
bb auth show
```

## 3. First commands

From inside a clone of a Bitbucket repository:

```sh
bb pr list          # all open pull requests
bb pr show 42       # comments on PR 42
bb pipeline latest  # most recent pipeline run
bb branch list      # all branches sorted by latest commit
bb browse           # open the repo in your browser
```

You don't need a checkout — `--project owner/repo` works from anywhere:

```sh
bb --project margus/bitbucket-cli pr list
```

## What's next

- [Commands reference](commands/) — every subcommand, with examples
- [Configuration](configuration.md) — env-var overrides, XDG paths
- [Shell completion](shell-completion.md) — tab-complete on bash, zsh, fish
