# bb-cli

[![CI](https://github.com/margus/bb-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/margus/bb-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/margus/bb-cli)](https://github.com/margus/bb-cli/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/margus/bb-cli.svg)](https://pkg.go.dev/github.com/margus/bb-cli)
[![License](https://img.shields.io/github/license/margus/bb-cli)](LICENSE)

Bitbucket Cloud REST API CLI. Pure Go stdlib, single static binary.
Config lives at `~/.config/bb/config.json` (or `$XDG_CONFIG_HOME/bb/config.json`).

## Install

Pre-built binaries (recommended):

```sh
# macOS arm64
curl -L -o /usr/local/bin/bb https://github.com/margus/bb-cli/releases/latest/download/bb-darwin-arm64
chmod +x /usr/local/bin/bb

# Linux amd64
curl -L -o /usr/local/bin/bb https://github.com/margus/bb-cli/releases/latest/download/bb-linux-amd64
chmod +x /usr/local/bin/bb
```

From source:

```sh
go install github.com/margus/bb-cli/cmd/bb@latest
```

Or build locally:

```sh
make build      # produces ./bb
make install    # go install to $GOPATH/bin
```

Requires Go 1.23+.

## Usage

```sh
bb auth                   # save Bitbucket username + app password
bb pr list                # list open PRs in the current repo
bb pr show 42             # show comments for PR 42
bb pr show 42 true        # show only unresolved inline comments
bb pr create main         # create PR from current branch into main
bb pipeline latest        # show latest pipeline
bb pipeline wait          # block until the latest pipeline completes
bb branch user margus     # branches whose author matches "margus"
bb env list               # deployment environments
bb browse                 # open the repo in the default browser
```

Work against an arbitrary repo without cd-ing into it:

```sh
bb --project owner/repo pr list
bb --project https://bitbucket.org/owner/repo branch list
```

`bb help` lists actions; `bb <action> help` lists methods and aliases.

## Actions

| Action       | Methods (aliases)                                                                                                                                                              |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `auth`       | `save`, `show`                                                                                                                                                                 |
| `branch`     | `list`/`l`, `user`/`u`, `name`/`n`                                                                                                                                             |
| `browse`     | `browse`/`b`, `show`/`url`                                                                                                                                                     |
| `env`        | `list`/`l`, `variables`/`v`, `create-variable`/`c`, `update-variable`/`u`                                                                                                      |
| `pipeline`   | `latest`, `get`, `wait`, `run`, `custom`/`c`                                                                                                                                   |
| `pr`         | `list`/`l`, `diff`/`d`, `files`, `commits`/`c`, `approve`/`a`, `no-approve`/`na`, `request-changes`/`rc`, `no-request-changes`/`nrc`, `decline`, `merge`/`m`, `create`, `show` |
| `pr-details` | `show`                                                                                                                                                                         |
| `upgrade`    | self-update from the GitHub Releases page                                                                                                                                      |

## Global flags

```text
--project <repo>     work against this repo instead of the cwd's git origin
-i, --interactive    prompt for PR title/description on `pr create`
--title <title>      set PR title for `pr create`
--description <txt>  set PR description for `pr create`
```

## Config

`~/.config/bb/config.json`:

```json
{
  "auth": {
    "username": "yourname",
    "appPassword": "abc..."
  }
}
```

Get an app password at <https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/>.

## Releases

Tagged builds are produced by goreleaser via `.github/workflows/release.yml`.
Pushing a tag matching `v*` triggers a release that uploads binaries named
`bb-<os>-<arch>` (plus `checksums.txt`) — the same names `bb upgrade` reads.

To cut a release:

```sh
git tag v0.1.0
git push origin v0.1.0
```

Local dry run (requires `goreleaser` installed):

```sh
goreleaser release --snapshot --clean
```

## Development

```sh
make build         # compile ./bb
make test          # go test -race ./...
make vet           # go vet ./...
make snapshot      # local goreleaser dry-run (produces dist/)
```

Lint (matches CI): `golangci-lint run`.

## Notes

- `NO_COLOR` env var disables ANSI color output.
- `bb upgrade` downloads `bb-<goos>-<goarch>` from the latest GitHub release.
- `bb version` reports the build version, short commit, and build date.
- Layout: `cmd/bb` entrypoint, `internal/{actions,api,config,util}`.

## License

[MIT](LICENSE).
