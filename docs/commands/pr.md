# bb pr

Pull request lifecycle — create, review, comment, approve, merge.

## Subcommands

| Command | Purpose | Aliases |
|---|---|---|
| `bb pr list [destination-branch]` | List open PRs (optionally filtered by dest branch) | `l` |
| `bb pr show <pr-id> [unresolved]` | Display all comments (general + inline). Pass `true` to filter to unresolved inline only. | |
| `bb pr diff <pr-id>` | Print unified diff | `d` |
| `bb pr files <pr-id>` | List changed files | |
| `bb pr commits <pr-id>` | List commit messages | `c` |
| `bb pr create <from> <to> [add-default-reviewers]` | Create a PR. With one arg, source is current git HEAD. | |
| `bb pr approve <pr-id>...` | Approve. Use `0` to approve all open. | `a` |
| `bb pr no-approve <pr-id>` | Revert an approval | `na` |
| `bb pr request-changes <pr-id>` | Request changes | `rc` |
| `bb pr no-request-changes <pr-id>` | Revert a change-request review | `nrc` |
| `bb pr decline <pr-id>` | Decline | |
| `bb pr merge <pr-id>` | Merge | `m` |
| `bb pr comment <pr-id> <message>` | Post a general comment | |
| `bb pr comment-inline <pr-id> <file> <line> <message>` | Post an inline review comment | |
| `bb pr checkout <pr-id>` | git fetch + git checkout the PR's source branch | |
| `bb pr view <pr-id>` | Open the PR in your browser | |

`bb pr` with no subcommand runs `list`.

## Examples

```sh
$ bb pr list
ID: 42
Author: alice
Source: feature/foo
Destination: main
Link: https://bitbucket.org/acme/widgets/pull-requests/42
Reviewers: Bob, Carol
Participants: Bob -> approved

$ bb pr show 42
$ bb pr show 42 true            # only unresolved inline comments

$ bb pr create feature/foo main
$ bb pr create main             # source = git symbolic-ref HEAD
$ bb --title "Hotfix" --description "Fixes outage" pr create feature/fix main
$ bb -i pr create feature/fix main    # interactive title/description prompt

$ bb pr approve 42
$ bb pr approve 0               # approve every open PR (use with care)

$ bb pr comment 42 "LGTM, thanks!"
$ bb pr comment-inline 42 path/to/file.go 25 "consider splitting this"

$ bb pr checkout 42             # equivalent to: git fetch origin feature/foo && git checkout feature/foo
$ bb pr view 42                 # opens https://bitbucket.org/.../pull-requests/42
$ bb pr merge 42
```

## `pr create` multi-target

Pass a comma-separated destination to open the same PR against several
branches at once (typical hotfix workflow):

```sh
bb pr create hotfix/x main,develop,release/2.0
```

This results in three separate PRs.

## Notes

- `pr list` issues one extra request per PR to fetch reviewers /
  participants — paginated lists with many PRs will be slow.
- `pr create` uses default reviewers from the repo settings and filters
  out the current user. Pass `0` as the third arg to skip default
  reviewers entirely.
- `--output json|yaml` works for `pr list`. Other PR commands print
  raw API output or simple success messages.
