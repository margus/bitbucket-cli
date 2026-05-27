# Configuration

`bb` is configured by three layers, in order of precedence:

1. **Command-line flags** (`--project`, `--output`, `-i`, …)
2. **Environment variables** (`BB_*`)
3. **Config file** at `~/.config/bb/config.json`

Each layer overrides the next. This is standard [viper](https://github.com/spf13/viper) behavior.

## Config file

Path: `$XDG_CONFIG_HOME/bb/config.json` if `XDG_CONFIG_HOME` is set, else `~/.config/bb/config.json`.

Permissions: `0600`. The file is created on first `bb auth token` or `bb auth save`.

Schema:

```json
{
  "auth": {
    "accessToken": "ATBB-...",
    "username": "alice",
    "appPassword": "..."
  }
}
```

Empty fields are persisted explicitly so switching auth methods doesn't leave stale credentials behind.

## Environment variables

| Variable | Maps to | Example |
|---|---|---|
| `BB_AUTH_ACCESSTOKEN` | `auth.accessToken` | `ATBB-xxxx...` |
| `BB_AUTH_USERNAME` | `auth.username` | `alice` |
| `BB_AUTH_APPPASSWORD` | `auth.appPassword` | `...` |
| `XDG_CONFIG_HOME` | base for config dir | `~/.config` |
| `NO_COLOR` | disable ANSI color | any value disables |

In CI, the typical pattern is to set just `BB_AUTH_ACCESSTOKEN` and skip the file entirely:

```yaml
# .github/workflows/example.yml
- name: List open PRs
  env:
    BB_AUTH_ACCESSTOKEN: ${{ secrets.BITBUCKET_TOKEN }}
  run: bb --project margus/bitbucket-cli pr list
```

## Global flags

These apply to every command:

| Flag | Description |
|---|---|
| `--project owner/repo` | act on this repo without needing a checkout |
| `-o, --output {json,yaml}` | machine-readable output (default: human-readable) |
| `--debug` | log every HTTP exchange to stderr |
| `-i, --interactive` | prompt for `--title` / `--description` on `pr create` |
| `--title <str>` | non-interactive PR title for `pr create` |
| `--description <str>` | non-interactive PR description for `pr create` |
| `--help` / `-h` | help for any command |
| `--version` / `-v` | print version + commit + build date |
