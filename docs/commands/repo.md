# bb repo

Inspect repositories in a workspace without needing a checkout.

## Subcommands

| Command | Purpose | Aliases |
|---|---|---|
| `bb repo list [workspace]` | List repositories | `l` |

If `[workspace]` is omitted, the workspace is taken from `--project`
(or from the git remote of the cwd).

## Examples

```sh
$ bb repo list acme
Slug: widgets
Name: Widgets
Language: go
Visibility: private

Slug: forms
Name: Forms
Language: php
Visibility: public

$ bb repo list   # workspace inferred from current repo or --project

$ bb --output json repo list acme | jq '.[] | select(.language == "go") | .slug'
"widgets"
```

## Notes

- Pagination is automatic, up to 100 pages.
- `--output json|yaml` is the recommended way to script over the result.
