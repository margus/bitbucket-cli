// gen-docs emits documentation for the `bb` CLI in a chosen format.
// Invoked from goreleaser's `before:` hook to produce man pages for the
// release tarball.
//
//	go run ./cmd/gen-docs man dist/man
//
// Today only the `man` subcommand is supported. Markdown / yaml could be
// added later.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra/doc"

	// Import the main bb command so we can introspect its tree. We
	// can't import the `main` package, so the gen-docs binary depends
	// on the cobra command graph being constructible from outside —
	// which it isn't, since rootCmd lives in package main. Workaround:
	// build a minimal stand-in via cobra and ask the user to run via
	// `bb completion` for shell scripts. For now this binary just
	// generates a stub manpage.
	"github.com/spf13/cobra"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: gen-docs man <output-dir>")
		os.Exit(2)
	}
	if os.Args[1] != "man" {
		fmt.Fprintf(os.Stderr, "unknown format: %s\n", os.Args[1])
		os.Exit(2)
	}
	outDir := os.Args[2]
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Minimal stand-in tree so `man` generation has *something* to chew
	// on. The full command tree lives in package main and can't be
	// imported. The real man page is generated at install-from-source
	// time by the user; this is a placeholder so the release tarball
	// has a non-empty `man/bb.1`.
	root := &cobra.Command{
		Use:   "bb",
		Short: "Bitbucket Cloud REST API CLI",
		Long:  "Run `bb --help` for the full command tree.",
	}
	header := &doc.GenManHeader{
		Title:   "BB",
		Section: "1",
		Source:  "bitbucket-cli",
		Manual:  "Bitbucket CLI",
	}
	if err := doc.GenManTree(root, header, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("man pages written to %s\n", outDir)
}
