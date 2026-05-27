# Authentication

`bb` supports two Bitbucket auth mechanisms. When both are configured,
the access token wins — the CLI sends `Authorization: Bearer <token>`
and ignores the app password.

## Access token (recommended)

Modern Bitbucket access tokens are scoped, revocable, and Atlassian's
official recommendation for CLI/CI usage.

### Create one

In the Bitbucket web UI:

- **Workspace token** (broadest scope): *Workspace settings → Access tokens → Create token*
- **Repository token** (single repo): *Repository settings → Access tokens → Create token*
- **Project token** (one Bitbucket project): *Project settings → Access tokens → Create token*

Grant the scopes you need. Common combinations:

| Use case | Scopes |
|---|---|
| Read-only browsing (`pr list`, `pipeline latest`, `branch list`) | `repository`, `pullrequest`, `pipeline` |
| Create / approve / comment on PRs | `pullrequest:write` |
| Trigger pipelines | `pipeline:write` |
| Manage environment variables | `pipeline:write` |

### Store it in bb

```sh
bb auth token
# pastes are masked from your shell history if you use `read -s` style
```

The CLI saves it to `~/.config/bb/config.json` with `0600` permissions.

## Username + app password (legacy)

For older setups that still need HTTP Basic auth:

```sh
bb auth save
```

App password docs:
<https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/>.

## Environment variables

For CI or scripted use, set credentials in env vars instead of writing
to disk. Env vars override file values (viper precedence: flag > env >
file > default):

```sh
export BB_AUTH_ACCESSTOKEN=ATBB-xxxxxxxxxxxxxxxxxxxx
# or
export BB_AUTH_USERNAME=alice
export BB_AUTH_APPPASSWORD=...
```

## Switching, clearing, inspecting

```sh
bb auth show     # which method is active + the stored value
bb auth logout   # wipe both token and app password
```

Running `bb auth token` or `bb auth save` always overwrites *both*
credentials, so switching methods doesn't leave stale data on disk.

## Where it's stored

```
$XDG_CONFIG_HOME/bb/config.json     # if XDG_CONFIG_HOME is set
~/.config/bb/config.json            # otherwise
```

File mode: `0600` (owner read/write only).
