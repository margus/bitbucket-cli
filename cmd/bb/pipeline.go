package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/margus/bitbucket-cli/internal/util"
	"github.com/spf13/cobra"
)

// pipelineWaitSleep is the polling interval used by `bb pipeline wait`.
// Package-level so tests can drop it to zero.
var pipelineWaitSleep = 2 * time.Second

var pipelineCmd = &cobra.Command{
	Use:               "pipeline",
	Short:             "Run, inspect, and wait on Bitbucket pipelines",
	PersistentPreRunE: requireGit(),
	RunE: func(cmd *cobra.Command, args []string) error {
		return pipelineLatest()
	},
}

var pipelineGetCmd = &cobra.Command{
	Use:   "get <pipeline-number>",
	Short: "Show details of a specific pipeline",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := pipelineGet(args[0], false)
		return err
	},
}

var pipelineLatestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Show details of the latest pipeline",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pipelineLatest()
	},
}

var pipelineWaitCmd = &cobra.Command{
	Use:   "wait [pipeline-number]",
	Short: "Block until a pipeline completes (defaults to the latest)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		num := ""
		if len(args) > 0 {
			num = args[0]
		}
		return pipelineWait(num)
	},
}

var pipelineRunCmd = &cobra.Command{
	Use:   "run <branch>",
	Short: "Trigger a pipeline run for the given branch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return pipelineRun(args[0])
	},
}

var pipelineLogsCmd = &cobra.Command{
	Use:   "logs <pipeline-number> [step-uuid]",
	Short: "Print the log output of a pipeline (or a specific step within it)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		step := ""
		if len(args) == 2 {
			step = args[1]
		}
		return pipelineLogs(args[0], step)
	},
}

var pipelineCustomCmd = &cobra.Command{
	Use:     "custom <branch> <pipeline-name>",
	Aliases: []string{"c"},
	Short:   "Trigger a custom pipeline (defined in bitbucket-pipelines.yml) on the given branch",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return pipelineCustom(args[0], args[1])
	},
}

type pipelineDetail struct {
	Creator struct {
		DisplayName string `json:"display_name"`
	} `json:"creator"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	Target struct {
		RefName string `json:"ref_name"`
	} `json:"target"`
	State struct {
		Name   string `json:"name"`
		Result struct {
			Name string `json:"name"`
		} `json:"result"`
	} `json:"state"`
	CreatedOn   string `json:"created_on"`
	CompletedOn string `json:"completed_on"`
	BuildNumber int    `json:"build_number"`
}

type pipelineSummary struct {
	ID          string `json:"id" yaml:"id"`
	Creator     string `json:"creator" yaml:"creator"`
	Repository  string `json:"repository" yaml:"repository"`
	Target      string `json:"target" yaml:"target"`
	State       string `json:"state" yaml:"state"`
	StateResult string `json:"stateResult" yaml:"stateResult"`
	Created     string `json:"created" yaml:"created"`
	Completed   string `json:"completed" yaml:"completed"`
	Link        string `json:"link" yaml:"link"`
}

func pipelineGet(pipelineNumber string, returnOnly bool) (*pipelineDetail, error) {
	c := apiClient()
	var resp pipelineDetail
	if err := c.JSON("GET", "/pipelines/"+pipelineNumber, nil, true, &resp); err != nil {
		return nil, err
	}
	if returnOnly {
		return &resp, nil
	}
	repo, _ := util.RepoPath()
	summary := pipelineSummary{
		ID:          pipelineNumber,
		Creator:     resp.Creator.DisplayName,
		Repository:  resp.Repository.Name,
		Target:      resp.Target.RefName,
		State:       resp.State.Name,
		StateResult: resp.State.Result.Name,
		Created:     resp.CreatedOn,
		Completed:   resp.CompletedOn,
		Link:        fmt.Sprintf("https://bitbucket.org/%s/addon/pipelines/home#!/results/%s", repo, pipelineNumber),
	}
	if render(summary) {
		return &resp, nil
	}
	util.O(map[string]string{
		"id":          pipelineNumber,
		"creator":     resp.Creator.DisplayName,
		"repository":  resp.Repository.Name,
		"target":      resp.Target.RefName,
		"state":       resp.State.Name,
		"stateResult": resp.State.Result.Name,
		"created":     resp.CreatedOn,
		"completed":   resp.CompletedOn,
		"link":        fmt.Sprintf("https://bitbucket.org/%s/addon/pipelines/home#!/results/%s", repo, pipelineNumber),
	}, "yellow")
	return &resp, nil
}

func pipelineLatest() error {
	num, err := latestPipelineID()
	if err != nil {
		return err
	}
	_, err = pipelineGet(strconv.Itoa(num), false)
	return err
}

func pipelineWait(pipelineNumber string) error {
	if pipelineNumber == "" {
		num, err := latestPipelineID()
		if err != nil {
			return err
		}
		pipelineNumber = strconv.Itoa(num)
		util.O("Pipeline: "+pipelineNumber, "yellow")
	}
	for {
		resp, err := pipelineGet(pipelineNumber, true)
		if err != nil {
			return err
		}
		if resp.State.Name == "COMPLETED" {
			fmt.Println()
			_, err := pipelineGet(pipelineNumber, false)
			return err
		}
		util.ORaw(".", "yellow")
		time.Sleep(pipelineWaitSleep)
	}
}

func pipelineRun(branch string) error {
	c := apiClient()
	payload := map[string]any{
		"target": map[string]any{
			"ref_type": "branch",
			"type":     "pipeline_ref_target",
			"ref_name": branch,
		},
	}
	b, err := c.Request("POST", "/pipelines/", payload, true)
	if err != nil {
		return err
	}
	var pretty any
	if json.Unmarshal(b, &pretty) == nil {
		util.O(pretty, "yellow")
	} else {
		fmt.Println(string(b))
	}
	return nil
}

func pipelineCustom(branch, pipeline string) error {
	c := apiClient()
	payload := map[string]any{
		"target": map[string]any{
			"ref_type": "branch",
			"type":     "pipeline_ref_target",
			"ref_name": branch,
			"selector": map[string]any{
				"type":    "custom",
				"pattern": pipeline,
			},
		},
	}
	var resp pipelineDetail
	if err := c.JSON("POST", "/pipelines/", payload, true, &resp); err != nil {
		return err
	}
	repo, _ := util.RepoPath()
	util.O(map[string]string{
		"link": fmt.Sprintf("https://bitbucket.org/%s/addon/pipelines/home#!/results/%d", repo, resp.BuildNumber),
	}, "yellow")
	return nil
}

func pipelineLogs(pipelineNumber, stepUUID string) error {
	c := apiClient()
	// If no step is given, list the pipeline's steps and stream the log
	// of each in order. Single-step output is the common case so make
	// that path direct.
	if stepUUID != "" {
		body, err := c.Request("GET", "/pipelines/"+pipelineNumber+"/steps/"+stepUUID+"/log", nil, true)
		if err != nil {
			return err
		}
		fmt.Print(string(body))
		return nil
	}
	var resp struct {
		Values []struct {
			UUID string `json:"uuid"`
			Name string `json:"name"`
		} `json:"values"`
	}
	if err := c.JSON("GET", "/pipelines/"+pipelineNumber+"/steps/", nil, true, &resp); err != nil {
		return err
	}
	for _, step := range resp.Values {
		util.O(fmt.Sprintf("--- step: %s (%s) ---", step.Name, step.UUID), "cyan")
		body, err := c.Request("GET", "/pipelines/"+pipelineNumber+"/steps/"+step.UUID+"/log", nil, true)
		if err != nil {
			return err
		}
		fmt.Print(string(body))
		fmt.Println()
	}
	return nil
}

func latestPipelineID() (int, error) {
	c := apiClient()
	var resp struct {
		Size int `json:"size"`
	}
	if err := c.JSON("GET", "/pipelines/", nil, true, &resp); err != nil {
		return 0, err
	}
	return resp.Size, nil
}

func init() {
	pipelineCmd.AddCommand(pipelineGetCmd, pipelineLatestCmd, pipelineWaitCmd, pipelineRunCmd, pipelineCustomCmd, pipelineLogsCmd)
	rootCmd.AddCommand(pipelineCmd)
}
