# Contributing

Thanks for considering a contribution.

## Reporting bugs

File an issue at <https://github.com/margus/bitbucket-cli/issues> using
the bug-report template. Helpful information to include:

- `bb version` output
- The exact command you ran (with sensitive arguments redacted)
- Stderr output, especially with `--debug` enabled
- Expected vs actual behavior

For security issues, follow [`SECURITY.md`](SECURITY.md) instead.

## Suggesting features

Open an issue with the feature-request template. Lay out:

- What you're trying to accomplish
- Why the existing commands don't already cover it
- What the proposed UX would look like (`bb …`)

## Pull requests

1. Fork the repo, create a topic branch.
2. Make the change. Keep commits focused.
3. `make ci-local` — runs vet, build, race tests, lint, govulncheck,
   and goreleaser check. Must pass before review.
4. Coverage: new code should ship with tests. The project floor is 80%
   per-package — don't drop it.
5. Open a PR against `main`. The CI workflow runs the same checks on
   ubuntu/macos/windows.

### Local development

See [`docs/development.md`](docs/development.md) for layout, conventions,
and the testing patterns.

```sh
make build          # ./bb
make test           # race + coverage
make lint           # golangci-lint
make ci-local       # everything CI runs
```

### Commit style

- Imperative subject line, sentence case ("Add foo", not "added foo" or "adds foo").
- Body wraps at 72 columns, explains *why* the change is needed.
- One logical change per commit; reviewers shouldn't have to mentally
  separate "fix typo" from "rewrite auth flow".

### Code style

- Standard `gofmt` + `goimports` formatting; `make lint` enforces.
- Public symbols get a doc comment that starts with the symbol name.
- Subcommand files in `cmd/bb/` follow the existing pattern: one cobra
  command per `var`, implementation in lowercase helpers, tests in a
  sibling `_test.go` driving `rootCmd` end-to-end through httptest.

## License

By contributing, you agree your changes will be released under the same
[MIT license](LICENSE) as the rest of the project.
