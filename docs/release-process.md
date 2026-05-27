# Release process

Releases are cut by pushing a tag. GitHub Actions then runs goreleaser
which builds, signs, and publishes binaries + SBOM + checksums.

## Cutting a release

```sh
# 1. Make sure main is green
gh run list --workflow CI --branch main --limit 1

# 2. Tag and push
git tag v0.2.0
git push origin v0.2.0

# 3. Watch the release workflow
gh run watch -R margus/bitbucket-cli "$(gh run list --workflow release --branch v0.2.0 --json databaseId -q '.[0].databaseId')"
```

Goreleaser publishes:

| Artifact | Purpose |
|---|---|
| `bb-linux-amd64`, `bb-linux-arm64` | Linux binaries |
| `bb-darwin-amd64`, `bb-darwin-arm64` | macOS Intel + Apple Silicon |
| `bb-windows-amd64.exe` | Windows |
| `checksums.txt` | SHA-256 of every asset |
| `checksums.txt.pem` + `.sig` | sigstore/cosign keyless signatures |
| `bom.json` | CycloneDX SBOM |

After the release is live:
- The `bb upgrade` self-update path works for every existing install.
- The Homebrew formula in `margus/homebrew-tap` is automatically
  updated by goreleaser; users run `brew upgrade bb`.
- The Docker image at `ghcr.io/margus/bitbucket-cli:vX.Y.Z` (and `:latest`)
  is published from the release workflow.

## Snapshot (dry run)

To verify everything end-to-end without publishing:

```sh
make snapshot
ls dist/
```

`goreleaser --snapshot --clean` runs locally with `GITHUB_TOKEN=""` and
writes everything to `dist/`. Verify the asset names match the
`bb-<os>-<arch>` pattern that `bb upgrade` expects.

## Versioning

We use semver-ish tags: `vMAJOR.MINOR.PATCH`. Bump:

- **PATCH** for bug fixes
- **MINOR** for new commands or flags (backward compatible)
- **MAJOR** for breaking changes to subcommands, flags, or the config
  file format

The `version` string in the binary is set via `-ldflags` to the git tag.
`bb version` prints it alongside the short commit and build date.

## Verifying a release

Cosign keyless verification (no key file needed):

```sh
cosign verify-blob \
  --signature checksums.txt.sig \
  --certificate checksums.txt.pem \
  --certificate-identity-regexp 'https://github.com/margus/bitbucket-cli/' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt
```

Then verify the binary against the (now trusted) checksums:

```sh
shasum -a 256 -c checksums.txt --ignore-missing
```
