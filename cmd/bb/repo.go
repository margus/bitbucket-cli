package main

import (
	"fmt"
	"strings"

	"github.com/margus/bitbucket-cli/internal/util"
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Inspect repositories without needing a checkout",
}

var repoListCmd = &cobra.Command{
	Use:     "list [workspace]",
	Aliases: []string{"l"},
	Short:   "List repositories in a workspace (paginated)",
	Long: `List repositories in the given workspace, or in the workspace inferred
from --project (or the current git remote) when no argument is supplied.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := resolveWorkspace(args)
		if err != nil {
			return err
		}
		return repoList(ws)
	},
}

func resolveWorkspace(args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	// Fall back to the workspace half of the current --project / git remote.
	repo, err := util.RepoPath()
	if err != nil {
		return "", fmt.Errorf("no workspace given and could not derive one: %w", err)
	}
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		return "", fmt.Errorf("could not parse workspace from %q", repo)
	}
	return parts[0], nil
}

type repoEntry struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	IsPrivate bool   `json:"is_private"`
}

func repoList(workspace string) error {
	c := apiClient()
	url := "/repositories/" + workspace + "?pagelen=100"
	var all []repoEntry
	for page := 1; page <= 100; page++ {
		var resp struct {
			Values []repoEntry `json:"values"`
			Next   string      `json:"next"`
		}
		if err := c.JSON("GET", fmt.Sprintf("%s&page=%d", url, page), nil, false, &resp); err != nil {
			return err
		}
		all = append(all, resp.Values...)
		if resp.Next == "" {
			break
		}
	}
	if render(all) {
		return nil
	}
	for _, r := range all {
		vis := "public"
		if r.IsPrivate {
			vis = "private"
		}
		util.O(map[string]string{
			"slug":       r.Slug,
			"name":       r.Name,
			"language":   r.Language,
			"visibility": vis,
		}, "yellow")
	}
	return nil
}

func init() {
	repoCmd.AddCommand(repoListCmd)
	rootCmd.AddCommand(repoCmd)
}
