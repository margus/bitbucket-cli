# Development

`bb` is a pure-Go project. Go 1.23+, no other runtime deps.

## Build, test, lint

```sh
make build        # produces ./bb
make test         # go test -race ./...
make cover        # coverage report
make lint         # golangci-lint run
make snapshot     # local goreleaser dry-run (produces dist/)
```

Full CI run, locally:

```sh
make ci-local
```

## Layout

```
cmd/bb/             # cobra entrypoint and all subcommands
  root.go           # root cmd + persistent flags + apiClient indirection
  auth.go           # bb auth {token,save,logout,show}
  branch.go
  browse.go
  env.go
  pipeline.go
  pr.go             # the big one — 14 subcommands
  prdetails.go      # pr show implementation
  repo.go
  workspace.go
  upgrade.go
  output.go         # --output json/yaml render() helper
internal/
  api/              # HTTP client, retry logic, Bearer/Basic auth
  config/           # viper-backed config; auth load/save
  util/             # color output, prompts, repo helpers, time format
```

## Conventions

- One file per top-level cobra command group in `cmd/bb/`.
- Each command's `RunE` calls a `<command><action>()` helper that's
  testable independently.
- API requests go through `apiClient()` (a package-level var in
  `root.go`) so tests can substitute an httptest-backed client.
- Use `util.O()` for human-readable output, gated by a `render()` check
  when the command supports `--output json|yaml`.
- Errors that have already been rendered (e.g. "no auth configured")
  return the `errSilent` sentinel; the top-level `main` checks
  `errors.Is(err, errSilent)` and skips the trailing red message.

## Testing patterns

Drive cobra commands through `runCmd` / `runRepoCmd` helpers in
`cmd/bb/testhelpers_test.go`. Each test stands up an `httptest.Server`,
asserts request shape, returns canned JSON, and inspects the captured
stdout.

For interactive prompts, override `util.PromptFn`. For external
processes (`git`, browser opener), override the swappable package vars
(`prCheckoutGitExec`, `prViewOpener`).

## Running against a real Bitbucket account

```sh
export BB_AUTH_ACCESSTOKEN=$YOUR_TOKEN
./bb --project margus/bitbucket-cli pr list
```
