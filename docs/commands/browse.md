# bb browse

Open the current Bitbucket repository in your default browser.

## Subcommands

| Command | Purpose | Aliases |
|---|---|---|
| `bb browse` | Open `https://bitbucket.org/<owner>/<repo>` in the browser | `b` |
| `bb browse url` | Print the repository URL without opening anything | `show` |

## Examples

```sh
$ bb browse url
https://bitbucket.org/margus/bitbucket-cli

$ bb browse
# (opens the page in $BROWSER / xdg-open / open / start)
```

## Notes

- `bb` resolves the repo from `git remote.origin.url`, or from
  `--project` if you pass it explicitly.
- The opener picked per OS: `open` on macOS, `xdg-open` on Linux,
  `start` on Windows.
