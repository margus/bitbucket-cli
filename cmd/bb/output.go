package main

import (
	"encoding/json"
	"fmt"

	"go.yaml.in/yaml/v3"
)

// Output formats. The default ("") keeps the existing human-readable
// rendering via util.O.
const (
	outputHuman = ""
	outputJSON  = "json"
	outputYAML  = "yaml"
)

// flagOutput is bound to the --output/-o persistent flag in root.go.
var flagOutput string

// render checks the --output flag. When it's "json" or "yaml", marshals
// data and prints it; returns true. When it's the default human output,
// returns false so the caller falls back to its util.O rendering.
func render(data any) bool {
	switch flagOutput {
	case outputJSON:
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Println("{}")
			return true
		}
		fmt.Println(string(b))
		return true
	case outputYAML:
		b, err := yaml.Marshal(data)
		if err != nil {
			fmt.Println("")
			return true
		}
		fmt.Print(string(b))
		return true
	default:
		return false
	}
}

// validateOutput returns an error if --output got an unknown value.
// Called from rootCmd.PersistentPreRunE so it fires once per invocation.
func validateOutput() error {
	switch flagOutput {
	case outputHuman, outputJSON, outputYAML:
		return nil
	}
	return fmt.Errorf("invalid --output %q (want one of: json, yaml)", flagOutput)
}
