package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/margus/bitbucket-cli/internal/actions"
	"github.com/margus/bitbucket-cli/internal/util"
)

// Build metadata. Populated via -ldflags at release time (goreleaser does
// this automatically). Dev builds keep the placeholder values.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var registry = map[string]actions.Action{
	"pr":         actions.Pr{},
	"pr-details": actions.PrDetails{},
	"pipeline":   actions.Pipeline{},
	"branch":     actions.Branch{},
	"auth":       actions.Auth{},
	"browse":     actions.Browse{},
	"env":        actions.Env{},
}

func main() {
	// Surface upgrade with the current version baked in.
	registry["upgrade"] = actions.Upgrade{CurrentVersion: version}

	gf, args := parseGlobalFlags(os.Args[1:])
	util.ProjectURL = gf.Project
	actions.PRTitle = gf.Title
	actions.PRDescription = gf.Description
	actions.Interactive = gf.Interactive

	switch {
	case len(args) == 0, isHelp(args[0]):
		printHelp()
		return
	case isVersion(args[0]):
		util.O("Version: "+version, "green")
		util.O("Commit:  "+commit, "gray")
		util.O("Built:   "+date, "gray")
		return
	case isAutocomplete(args[0]):
		printActionsForAutocomplete()
		return
	}

	actionName := args[0]
	a, ok := registry[actionName]
	if !ok {
		util.O(fmt.Sprintf("Given action is invalid: (%s)", actionName), "red")
		os.Exit(1)
	}

	// Sub-command.
	var subCmd string
	var rest []string
	if len(args) >= 2 {
		subCmd = args[1]
		rest = args[2:]
	}

	switch {
	case isHelp(subCmd):
		printActionHelp(a)
		return
	case isAutocomplete(subCmd):
		printSubcommandsForAutocomplete(a)
		return
	}

	// Git-folder check, skipped when --project is provided.
	if a.RequireGit() && util.ProjectURL == "" {
		if !util.HasGitDir() {
			util.O("ERROR: No git repository found in current directory.", "red")
			util.O(`Use --project "owner/repo" to work with remote repository.`, "yellow")
			os.Exit(1)
		}
	}

	method := a.Default()
	if subCmd != "" {
		method = actions.MethodFromAlias(a, subCmd)
		if method == "" {
			util.O(fmt.Sprintf("Unknown method %s for %s action.", subCmd, actionName), "red")
			os.Exit(1)
		}
	}

	if err := a.Dispatch(method, rest); err != nil {
		util.O(err.Error(), "red")
		os.Exit(1)
	}
}

// globalFlags holds the values of the cross-action flags pulled out before
// dispatch. Kept as a struct so parseGlobalFlags is a pure function that
// can be unit-tested without touching package globals.
type globalFlags struct {
	Project     string
	Title       string
	Description string
	Interactive bool
}

func parseGlobalFlags(argv []string) (globalFlags, []string) {
	var gf globalFlags
	out := make([]string, 0, len(argv))
	for i := 0; i < len(argv); i++ {
		switch argv[i] {
		case "--project":
			if i+1 < len(argv) {
				gf.Project = argv[i+1]
				i++
			}
		case "--title":
			if i+1 < len(argv) {
				gf.Title = argv[i+1]
				i++
			}
		case "--description":
			if i+1 < len(argv) {
				gf.Description = argv[i+1]
				i++
			}
		case "-i", "--interactive":
			gf.Interactive = true
		default:
			out = append(out, argv[i])
		}
	}
	return gf, out
}

func isHelp(s string) bool {
	return s == "help" || s == "--help" || s == "-h"
}
func isVersion(s string) bool {
	return s == "version" || s == "--version" || s == "-v"
}
func isAutocomplete(s string) bool {
	return s == "autocomplete" || s == "--autocomplete"
}

func printHelp() {
	util.O("Available actions:", "green")
	names := actionNames()
	for _, n := range names {
		util.O("  "+n, "green")
	}
	util.O("", "white")
	util.O("Global options:", "green")
	util.O(`  --project <repo>          Work with repository (e.g., --project "owner/repo" or --project "https://bitbucket.org/owner/repo")`, "yellow")
	util.O(`  -i, --interactive         Enable interactive mode (e.g., prompt for PR title and description)`, "yellow")
	util.O(`  --title <title>           Set PR title (used with pr create)`, "yellow")
	util.O(`  --description <desc>      Set PR description (used with pr create)`, "yellow")
}

func printActionHelp(a actions.Action) {
	util.O("Available methods:", "green")
	type entry struct{ method, aliases string }
	var rows []entry
	for m, al := range a.Commands() {
		rows = append(rows, entry{m, al})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].method < rows[j].method })
	for _, r := range rows {
		parts := strings.Split(r.aliases, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) == 1 {
			util.O(parts[0], "yellow")
			continue
		}
		// Mimic colored split (yellow first, gray rest) via direct ANSI.
		fmt.Print("\033[0;33m" + parts[0])
		for _, p := range parts[1:] {
			fmt.Print("\033[0;90m," + p)
		}
		fmt.Println("\033[0m")
	}
}

func actionNames() []string {
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func printActionsForAutocomplete() {
	fmt.Print(strings.Join(actionNames(), " "))
}

func printSubcommandsForAutocomplete(a actions.Action) {
	fmt.Print(strings.Join(actions.PrimaryAliases(a), " "))
}
