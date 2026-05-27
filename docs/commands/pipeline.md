# bb pipeline

Run, inspect, and wait on Bitbucket pipelines.

## Subcommands

| Command | Purpose | Aliases |
|---|---|---|
| `bb pipeline latest` | Details of the most recent pipeline | |
| `bb pipeline get <number>` | Details of pipeline `<number>` | |
| `bb pipeline wait [number]` | Block until pipeline completes (default: latest) | |
| `bb pipeline run <branch>` | Trigger the default pipeline on `<branch>` | |
| `bb pipeline custom <branch> <pipeline-name>` | Trigger a custom pipeline | `c` |
| `bb pipeline logs <number> [step-uuid]` | Print log output | |

`bb pipeline` with no subcommand runs `latest`.

## Examples

```sh
$ bb pipeline latest
ID: 123
Creator: Alice
Repository: widgets
Target: main
State: COMPLETED
StateResult: SUCCESSFUL
Created: 2026-05-26T14:00:00Z
Completed: 2026-05-26T14:04:31Z
Link: https://bitbucket.org/acme/widgets/addon/pipelines/home#!/results/123

$ bb pipeline run main           # trigger default branch pipeline
$ bb pipeline custom main deploy # trigger the "deploy" custom pipeline on main
$ bb pipeline wait               # tail the latest pipeline until it finishes
$ bb pipeline logs 123           # print every step's log, separated by headers
$ bb pipeline logs 123 {step-uuid}  # just that step
```

## Notes

- `wait` polls every 2 seconds. The interval is a package-level var
  (`pipelineWaitSleep`); tests set it to 0.
- `logs` without a step uuid lists steps first, then prints each step's
  output sequentially. Useful for `bb pipeline logs $(bb pipeline latest --output json | jq -r .id) | tee build.log`.
