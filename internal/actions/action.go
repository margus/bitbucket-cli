package actions

import (
	"sort"
	"strings"
)

// Action is the interface every command group implements.
type Action interface {
	// Default returns the method name to run when the user types just the
	// action with no subcommand (e.g. `bb pr` -> "list").
	Default() string
	// Commands maps method name -> comma-separated alias list
	// (e.g. "list" => "list, l").
	Commands() map[string]string
	// RequireGit returns true if the command needs to be run inside a git
	// repo (skipped when --project is given).
	RequireGit() bool
	// Dispatch runs the named method with the remaining positional args.
	Dispatch(method string, args []string) error
}

// MethodFromAlias resolves an alias like "l" or "list" to its method name.
// Returns "" if no method owns the alias.
func MethodFromAlias(a Action, alias string) string {
	for method, aliases := range a.Commands() {
		for _, x := range strings.Split(aliases, ",") {
			if strings.TrimSpace(x) == alias {
				return method
			}
		}
	}
	return ""
}

// PrimaryAliases returns the first alias of each command, sorted. Used by
// the autocomplete static command.
func PrimaryAliases(a Action) []string {
	out := make([]string, 0, len(a.Commands()))
	for _, aliases := range a.Commands() {
		out = append(out, strings.TrimSpace(strings.Split(aliases, ",")[0]))
	}
	sort.Strings(out)
	return out
}
