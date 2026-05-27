# bb upgrade

Self-update by fetching the latest GitHub release for your OS / arch
and replacing the running binary.

## Usage

```sh
bb upgrade
```

## What it does

1. `GET https://api.github.com/repos/margus/bitbucket-cli/releases/latest`
2. Compares the release tag to your current `bb version`. If your
   version is already at or above the latest, it exits with
   "You are already on the latest version".
3. Picks the asset matching `bb-<goos>-<goarch>` (or `bb-windows-amd64.exe`).
4. Downloads to a tempfile.
5. `chmod +x` then `os.Rename` over your current binary path
   (`os.Executable()`). Falls back to copy when the temp file and the
   target are on different filesystems.

## Example

```sh
$ bb version
bb v0.1.0
  commit: 935a970
  built:  2026-05-26

$ bb upgrade
Fetching new version (v0.2.0) ...
bb-cli updated

$ bb version
bb v0.2.0
  ...
```

## Notes

- The comparison is lexical string compare, not semver-aware. Tag your
  releases consistently (`v0.1.0` → `v0.2.0` → `v0.10.0` orders
  correctly, but `1.0.0` would mis-sort against `v9.9.9`).
- Verification: pair `bb upgrade` with `bb version` afterward.
- For a repeatable install — e.g. a versioned CI step — pull a specific
  asset directly:
  ```sh
  curl -L -o bb https://github.com/margus/bitbucket-cli/releases/download/v0.1.0/bb-linux-amd64
  ```
