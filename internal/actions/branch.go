package actions

import (
	"fmt"
	"strings"
	"time"

	"github.com/margus/bitbucket-cli/internal/api"
	"github.com/margus/bitbucket-cli/internal/util"
)

type Branch struct{}

func (Branch) Default() string { return "list" }
func (Branch) Commands() map[string]string {
	return map[string]string{
		"list": "list, l",
		"user": "user, u",
		"name": "name, n",
	}
}
func (Branch) RequireGit() bool { return true }

func (b Branch) Dispatch(method string, args []string) error {
	switch method {
	case "list":
		return b.list("", "")
	case "user":
		if len(args) < 1 {
			return fmt.Errorf("user requires a username argument")
		}
		return b.list(args[0], "")
	case "name":
		if len(args) < 1 {
			return fmt.Errorf("name requires a branch-name argument")
		}
		return b.list("", args[0])
	default:
		return fmt.Errorf("unknown method: %s", method)
	}
}

type branchListResponse struct {
	Values []struct {
		Name   string `json:"name"`
		Target struct {
			Date   string `json:"date"`
			Author struct {
				Raw  string `json:"raw"`
				User struct {
					DisplayName string `json:"display_name"`
				} `json:"user"`
			} `json:"author"`
		} `json:"target"`
	} `json:"values"`
	Next string `json:"next"`
}

func (Branch) list(user, branch string) error {
	c := api.New()
	type row struct {
		Branch  string `json:"branch"`
		User    string `json:"user"`
		Updated string `json:"updated"`
	}
	var rows []row

	page := 1
	for {
		var resp branchListResponse
		if err := c.JSON("GET", fmt.Sprintf("/refs/branches?page=%d", page), nil, true, &resp); err != nil {
			return err
		}
		for _, br := range resp.Values {
			owner := br.Target.Author.User.DisplayName
			if owner == "" {
				owner = br.Target.Author.Raw
			}
			if user != "" && !strings.Contains(strings.ToLower(owner), strings.ToLower(user)) {
				continue
			}
			if branch != "" && !strings.Contains(strings.ToLower(br.Name), strings.ToLower(branch)) {
				continue
			}
			updated := br.Target.Date
			if t, err := time.Parse(time.RFC3339, updated); err == nil {
				updated = t.Format("2006-01-02 15:04")
			}
			rows = append(rows, row{Branch: br.Name, User: owner, Updated: updated})
		}
		if resp.Next == "" {
			break
		}
		page++
	}

	for _, r := range rows {
		util.O(map[string]string{
			"branch":  r.Branch,
			"user":    r.User,
			"updated": r.Updated,
		}, "yellow")
	}
	return nil
}
