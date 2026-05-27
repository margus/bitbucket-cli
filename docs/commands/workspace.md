# bb workspace

List the Bitbucket workspaces the authenticated user has access to.

> **Implementation note**: Atlassian removed all cross-workspace listing
> endpoints (`/workspaces`, `/user/permissions/workspaces`,
> `/repositories?role=member`) as part of CHANGE-2770 (effective
> 2026-04-14). `bb` now calls **`GET /2.0/user/workspaces`**, the new
> replacement endpoint Atlassian released in January 2026 specifically
> for this use case.

## Subcommands

| Command | Purpose | Aliases |
|---|---|---|
| `bb workspace list` | List workspaces | `l` |

`bb workspace` with no subcommand runs `list`. The command alias `ws`
also works (`bb ws list`).

## Examples

```sh
$ bb workspace list
Slug: acme
Name: Acme Co
UUID: {a1b2-...}

Slug: widgets
Name: Widgets Inc
UUID: {c3d4-...}

$ bb --output json workspace list | jq -r '.[].slug'
acme
widgets
```

## Notes

- The result depends on the scope of the active token. A repository
  access token may only see the repo's workspace; a workspace access
  token sees that workspace; only OAuth flows with user scope see all
  workspaces.
