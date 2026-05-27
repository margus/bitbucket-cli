# bb auth

Manage Bitbucket credentials. See [`docs/auth.md`](../auth.md) for the
full conceptual guide.

## Subcommands

| Command | Purpose |
|---|---|
| `bb auth token` | Save an access token (recommended). Prompts you to paste. |
| `bb auth save` | Save username + app password (legacy). |
| `bb auth logout` | Clear both credentials. |
| `bb auth show` | Display the active method + stored value(s). |

## Examples

```sh
$ bb auth token
Paste a Bitbucket access token (workspace, repository, or project scope).
Create one at: https://support.atlassian.com/bitbucket-cloud/docs/access-tokens/
Token: ATBB-xxxxxxxx
Access token saved.

$ bb auth show
AccessToken: ATBB-xxxxxxxx
Method: access token

$ bb auth logout
Logged out.
```

## CI usage

Skip the file entirely. Set `BB_AUTH_ACCESSTOKEN` and `bb` picks it up:

```sh
export BB_AUTH_ACCESSTOKEN=$BITBUCKET_TOKEN
bb pr list
```
