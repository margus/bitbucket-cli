# Shell completion

`bb` ships generated completion scripts via cobra — bash, zsh, fish, and
PowerShell are all supported.

## bash

Linux:

```sh
bb completion bash > /etc/bash_completion.d/bb
```

macOS (Homebrew bash-completion):

```sh
bb completion bash > /usr/local/etc/bash_completion.d/bb     # Intel
bb completion bash > /opt/homebrew/etc/bash_completion.d/bb  # Apple Silicon
```

## zsh

```sh
bb completion zsh > "${fpath[1]}/_bb"
```

Or load only for the current shell:

```sh
source <(bb completion zsh)
```

If you've never enabled completion on zsh:

```sh
echo 'autoload -U compinit; compinit' >> ~/.zshrc
```

## fish

```sh
bb completion fish > ~/.config/fish/completions/bb.fish
```

Or for the current session only:

```sh
bb completion fish | source
```

## PowerShell

```powershell
bb completion powershell | Out-String | Invoke-Expression
```

To make it permanent, add that line to `$PROFILE`.

## What gets completed

- Subcommand names (`auth`, `pr`, `pipeline`, …)
- Subcommand aliases (`bb pr l` autocompletes via `list`)
- Flag names after `--`
- Static flag values (e.g. `--output ` suggests `json`, `yaml`)
