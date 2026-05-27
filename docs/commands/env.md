# bb env

Manage Bitbucket deployment environments and their variables.

## Subcommands

| Command | Purpose | Aliases |
|---|---|---|
| `bb env list` | List deployment environments | `l` |
| `bb env variables <env-uuid>` | List variables in an environment | `v` |
| `bb env create-variable <env-uuid> <key> <value> [secured]` | Create a variable | `c` |
| `bb env update-variable <env-uuid> <var-uuid> <key> <value> [secured]` | Update a variable | `u` |

`bb env` with no subcommand runs `list`. The `secured` argument is a
boolean (`true`/`false`/`1`/`0`); when true, the value is write-only.

## Examples

```sh
$ bb env list
UUID: {a1b2-...}
Name: production

UUID: {c3d4-...}
Name: staging

$ bb env variables {a1b2-...}
UUID: {var-1}
Key: API_KEY
Value:
Secured: Yes

$ bb env create-variable {a1b2-...} DEBUG 1 false
UUID: {var-2}
Key: DEBUG
Value: 1
Secured: No

$ bb env update-variable {a1b2-...} {var-2} DEBUG 0 false
```

## Notes

- Environment UUIDs include curly braces in the wire format — pass them
  verbatim, no shell quoting tricks needed.
- Secured variable values aren't returned by Bitbucket; the field is
  empty on read.
