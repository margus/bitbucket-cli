package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/margus/bitbucket-cli/internal/util"
	"github.com/spf13/cobra"
)

var branchCmd = &cobra.Command{
	Use:               "branch",
	Short:             "List repository branches",
	Aliases:           []string{"b"},
	PersistentPreRunE: requireGit(),
	RunE: func(cmd *cobra.Command, args []string) error {
		return branchList("", "")
	},
}

var branchListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List all branches in the current repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		return branchList("", "")
	},
}

var branchUserCmd = &cobra.Command{
	Use:     "user <username>",
	Aliases: []string{"u"},
	Short:   "List branches whose latest commit author matches <username>",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return branchList(args[0], "")
	},
}

var branchNameCmd = &cobra.Command{
	Use:     "name <substring>",
	Aliases: []string{"n"},
	Short:   "List branches whose name contains <substring>",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return branchList("", args[0])
	},
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

type branchRow struct {
	Branch  string `json:"branch" yaml:"branch"`
	User    string `json:"user" yaml:"user"`
	Updated string `json:"updated" yaml:"updated"`
}

func branchList(user, branch string) error {
	c := apiClient()
	var rows []branchRow
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
			rows = append(rows, branchRow{Branch: br.Name, User: owner, Updated: updated})
		}
		if resp.Next == "" {
			break
		}
		page++
	}

	if render(rows) {
		return nil
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

func init() {
	branchCmd.AddCommand(branchListCmd, branchUserCmd, branchNameCmd)
	rootCmd.AddCommand(branchCmd)
}
