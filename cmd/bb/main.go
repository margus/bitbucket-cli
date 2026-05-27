package main

import (
	"errors"
	"os"

	"github.com/margus/bitbucket-cli/internal/util"
)

// errSilent: subcommands that have already rendered their own
// human-readable error message return this so main exits non-zero
// without printing again.
var errSilent = errors.New("silent")

func main() {
	if err := rootCmd.Execute(); err != nil {
		if !errors.Is(err, errSilent) {
			util.O(err.Error(), "red")
		}
		os.Exit(1)
	}
}
