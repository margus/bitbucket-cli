package actions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/margus/bitbucket-cli/internal/api"
	"github.com/margus/bitbucket-cli/internal/util"
)

type Pipeline struct{}

func (Pipeline) Default() string { return "latest" }
func (Pipeline) Commands() map[string]string {
	return map[string]string{
		"get":    "get",
		"latest": "latest",
		"wait":   "wait",
		"run":    "run",
		"custom": "custom, c",
	}
}
func (Pipeline) RequireGit() bool { return true }

func (p Pipeline) Dispatch(method string, args []string) error {
	switch method {
	case "get":
		if len(args) < 1 {
			return fmt.Errorf("get requires <pipeline-number>")
		}
		_, err := p.get(args[0], false)
		return err
	case "latest":
		return p.latest()
	case "wait":
		num := ""
		if len(args) > 0 {
			num = args[0]
		}
		return p.wait(num)
	case "run":
		if len(args) < 1 {
			return fmt.Errorf("run requires <branch>")
		}
		return p.run(args[0])
	case "custom":
		if len(args) < 2 {
			return fmt.Errorf("custom requires <branch> <pipeline-name>")
		}
		return p.custom(args[0], args[1])
	default:
		return fmt.Errorf("unknown method: %s", method)
	}
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

func (Pipeline) get(pipelineNumber string, returnOnly bool) (*pipelineDetail, error) {
	c := api.New()
	var resp pipelineDetail
	if err := c.JSON("GET", "/pipelines/"+pipelineNumber, nil, true, &resp); err != nil {
		return nil, err
	}
	if returnOnly {
		return &resp, nil
	}
	repo, _ := util.RepoPath()
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

func (p Pipeline) latest() error {
	num, err := p.getLatestPipelineID()
	if err != nil {
		return err
	}
	_, err = p.get(strconv.Itoa(num), false)
	return err
}

func (p Pipeline) wait(pipelineNumber string) error {
	if pipelineNumber == "" {
		num, err := p.getLatestPipelineID()
		if err != nil {
			return err
		}
		pipelineNumber = strconv.Itoa(num)
		util.O("Pipeline: "+pipelineNumber, "yellow")
	}

	for {
		resp, err := p.get(pipelineNumber, true)
		if err != nil {
			return err
		}
		if resp.State.Name == "COMPLETED" {
			fmt.Println()
			_, err := p.get(pipelineNumber, false)
			return err
		}
		util.ORaw(".", "yellow")
		time.Sleep(2 * time.Second)
	}
}

func (Pipeline) run(branch string) error {
	c := api.New()
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
	// Pretty-print JSON to match `o($response)` for an array.
	var pretty any
	if err := json.Unmarshal(b, &pretty); err == nil {
		util.O(pretty, "yellow")
	} else {
		fmt.Println(string(b))
	}
	return nil
}

func (Pipeline) custom(branch, pipeline string) error {
	c := api.New()
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

func (Pipeline) getLatestPipelineID() (int, error) {
	c := api.New()
	var resp struct {
		Size int `json:"size"`
	}
	if err := c.JSON("GET", "/pipelines/", nil, true, &resp); err != nil {
		return 0, err
	}
	return resp.Size, nil
}
