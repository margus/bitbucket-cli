# bb branch

List branches in a Bitbucket repository, optionally filtered.

## Subcommands

| Command | Purpose | Aliases |
|---|---|---|
| `bb branch list` | List all branches (paginated, latest commit metadata) | `l` |
| `bb branch user <name>` | Filter to branches whose latest commit author matches `<name>` (substring, case-insensitive) | `u` |
| `bb branch name <substr>` | Filter to branches whose name contains `<substr>` | `n` |

`bb branch` with no subcommand is equivalent to `bb branch list`.

## Examples

```sh
$ bb branch list
Branch: main
User: Alice
Updated: 2026-05-22 14:00

Branch: feature/widgets
User: Bob
Updated: 2026-05-23 09:30

$ bb branch user alice    # all branches authored by Alice (or aliCE, etc.)
$ bb branch name feature  # all branches matching */feature*/*
```

## Notes

- Pagination is automatic — the CLI keeps following `next` links until
  the listing is exhausted.
- The author is taken from `target.author.user.display_name` if present,
  otherwise the raw `name <email>` string.
- Output respects `--output json|yaml`.
