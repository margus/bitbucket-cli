# Security Policy

## Reporting a vulnerability

If you find a security issue in `bitbucket-cli`, please **do not** open a
public GitHub issue. Instead, use GitHub's private security advisory:

- <https://github.com/margus/bitbucket-cli/security/advisories/new>

I'll respond within 7 days with either a fix timeline or a request for
more information. Coordinated disclosure is expected — please give me a
reasonable window (typically 30 days) before publishing details.

## Supported versions

Only the latest minor release receives security fixes. After a patch
release, the prior `v0.X.Y` is end-of-life.

| Version | Supported |
|---|---|
| latest minor | ✅ |
| older minors | ❌ |

## What counts as a security issue

- Credential exfiltration via crafted Bitbucket responses, env vars, or
  config files
- Privilege escalation, e.g. writing outside the `~/.config/bb/` dir
- Argument-handling bugs that let an attacker execute arbitrary commands
  (PR titles, branch names, etc. being passed unescaped to `exec`)
- Bundled dependency vulnerabilities flagged by `govulncheck`

## What's intentionally out of scope

- The CLI prints credentials in `bb auth show`. That's documented
  behavior, not a bug; protect the terminal.
- The CLI stores credentials in plaintext at `~/.config/bb/config.json`
  with `0600` permissions. We don't (yet) integrate with OS keychains.
- Running `bb` as root, in shared user accounts, or on multi-tenant
  hosts is your responsibility.

## Supply chain hardening

Each release artifact is:

- **Built by GitHub Actions** in a reproducible workflow (no external
  build hosts).
- **Signed** with sigstore/cosign keyless. Verify with the snippet in
  [`docs/release-process.md`](docs/release-process.md).
- **Accompanied by a SBOM** (CycloneDX, `bom.json`) and `checksums.txt`.
- **Scanned** in CI via `govulncheck` (Go stdlib + dep CVEs) and
  `dependency-review-action` (PRs).
