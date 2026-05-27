package main

import (
	"github.com/margus/bitbucket-cli/internal/util"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:               "env",
	Short:             "Manage deployment environments and their variables",
	PersistentPreRunE: requireGit(),
	RunE: func(cmd *cobra.Command, args []string) error {
		return envList()
	},
}

var envListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List deployment environments",
	RunE: func(cmd *cobra.Command, args []string) error {
		return envList()
	},
}

var envVariablesCmd = &cobra.Command{
	Use:     "variables <env-uuid>",
	Aliases: []string{"v"},
	Short:   "List variables for an environment",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return envVariables(args[0])
	},
}

var envCreateVarCmd = &cobra.Command{
	Use:     "create-variable <env-uuid> <key> <value> [secured]",
	Aliases: []string{"c"},
	Short:   "Create an environment variable",
	Args:    cobra.RangeArgs(3, 4),
	RunE: func(cmd *cobra.Command, args []string) error {
		secured := len(args) == 4 && parseBoolish(args[3])
		return envCreateVar(args[0], args[1], args[2], secured)
	},
}

var envUpdateVarCmd = &cobra.Command{
	Use:     "update-variable <env-uuid> <var-uuid> <key> <value> [secured]",
	Aliases: []string{"u"},
	Short:   "Update an environment variable",
	Args:    cobra.RangeArgs(4, 5),
	RunE: func(cmd *cobra.Command, args []string) error {
		secured := len(args) == 5 && parseBoolish(args[4])
		return envUpdateVar(args[0], args[1], args[2], args[3], secured)
	},
}

type envListResp struct {
	Values []envEntry `json:"values"`
}

type envEntry struct {
	UUID string `json:"uuid" yaml:"uuid"`
	Name string `json:"name" yaml:"name"`
}

func envList() error {
	c := apiClient()
	var r envListResp
	if err := c.JSON("GET", "/environments", nil, true, &r); err != nil {
		return err
	}
	if render(r.Values) {
		return nil
	}
	for _, env := range r.Values {
		util.O(map[string]string{"uuid": env.UUID, "name": env.Name}, "yellow")
	}
	return nil
}

type envVar struct {
	UUID    string `json:"uuid" yaml:"uuid"`
	Key     string `json:"key" yaml:"key"`
	Value   string `json:"value,omitempty" yaml:"value,omitempty"`
	Secured bool   `json:"secured" yaml:"secured"`
}

func envVariables(envUUID string) error {
	c := apiClient()
	var r struct {
		Values []envVar `json:"values"`
	}
	if err := c.JSON("GET", "/deployments_config/environments/"+envUUID+"/variables", nil, true, &r); err != nil {
		return err
	}
	if render(r.Values) {
		return nil
	}
	for _, v := range r.Values {
		printEnvVar(v)
	}
	return nil
}

func envCreateVar(envUUID, key, value string, secured bool) error {
	c := apiClient()
	var resp envVar
	payload := map[string]any{"key": key, "value": value, "secured": secured}
	if err := c.JSON("POST", "/deployments_config/environments/"+envUUID+"/variables", payload, true, &resp); err != nil {
		return err
	}
	if render(resp) {
		return nil
	}
	printEnvVar(resp)
	return nil
}

func envUpdateVar(envUUID, varUUID, key, value string, secured bool) error {
	c := apiClient()
	var resp envVar
	payload := map[string]any{"key": key, "value": value, "secured": secured}
	if err := c.JSON("PUT", "/deployments_config/environments/"+envUUID+"/variables/"+varUUID, payload, true, &resp); err != nil {
		return err
	}
	if render(resp) {
		return nil
	}
	printEnvVar(resp)
	return nil
}

func printEnvVar(v envVar) {
	util.O(map[string]string{
		"uuid":    v.UUID,
		"key":     v.Key,
		"value":   v.Value,
		"secured": yesNo(v.Secured),
	}, "yellow")
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func init() {
	envCmd.AddCommand(envListCmd, envVariablesCmd, envCreateVarCmd, envUpdateVarCmd)
	rootCmd.AddCommand(envCmd)
}
