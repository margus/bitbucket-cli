# bitbucket-cli

[![CI](https://github.com/margus/bitbucket-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/margus/bitbucket-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/margus/bitbucket-cli)](https://github.com/margus/bitbucket-cli/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/margus/bitbucket-cli.svg)](https://pkg.go.dev/github.com/margus/bitbucket-cli)
[![License](https://img.shields.io/github/license/margus/bitbucket-cli)](LICENSE)

Bitbucket Cloud REST API CLI written in Go. Single static binary, no
runtime dependencies. Manages pull requests, pipelines, branches,
deployment environments, workspaces, and repositories from your terminal.

## Install

Download the latest release binary for your platform:

```sh
# macOS arm64
curl -L -o /usr/local/bin/bb https://github.com/margus/bitbucket-cli/releases/latest/download/bb-darwin-arm64
chmod +x /usr/local/bin/bb

# Linux amd64
curl -L -o /usr/local/bin/bb https://github.com/margus/bitbucket-cli/releases/latest/download/bb-linux-amd64
chmod +x /usr/local/bin/bb
```

Or from source:

```sh
go install github.com/margus/bitbucket-cli/cmd/bb@latest
```

Other install paths (Homebrew, Docker, build from source) are in
[`docs/getting-started.md`](docs/getting-started.md).

## Quick start

```sh
bb auth token                  # paste a workspace access token from the Bitbucket UI
bb pr list                     # open PRs in the current repo
bb pr show 42                  # comments on PR 42
bb pipeline latest             # last pipeline run
```

## Documentation

- [Getting started](docs/getting-started.md) — install, first run, hello world
- [Authentication](docs/auth.md) — access tokens, app passwords, env vars
- [Configuration](docs/configuration.md) — config file, XDG, env-var overrides
- [Shell completion](docs/shell-completion.md) — bash / zsh / fish / powershell
- Commands:
  [`auth`](docs/commands/auth.md),
  [`branch`](docs/commands/branch.md),
  [`browse`](docs/commands/browse.md),
  [`env`](docs/commands/env.md),
  [`pipeline`](docs/commands/pipeline.md),
  [`pr`](docs/commands/pr.md),
  [`repo`](docs/commands/repo.md),
  [`workspace`](docs/commands/workspace.md),
  [`upgrade`](docs/commands/upgrade.md)
- [Development](docs/development.md) — building, testing, linting
- [Release process](docs/release-process.md) — tagging, goreleaser, bb upgrade

## License

[MIT](LICENSE).
