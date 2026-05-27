package main

import (
	"strconv"

	"github.com/margus/bitbucket-cli/internal/util"
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:     "workspace",
	Aliases: []string{"ws"},
	Short:   "List Bitbucket workspaces visible to the authenticated user",
	RunE: func(cmd *cobra.Command, args []string) error {
		return workspaceList()
	},
}

var workspaceListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List workspaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		return workspaceList()
	},
}

// workspaceRow is the projected shape we print. The wire shape from
// /user/workspaces wraps each entry in {type, administrator, workspace:
// {uuid, slug, links, ...}} — we flatten and only carry the fields that
// are actually useful at the CLI surface.
type workspaceRow struct {
	Slug          string `json:"slug" yaml:"slug"`
	UUID          string `json:"uuid" yaml:"uuid"`
	Administrator bool   `json:"administrator" yaml:"administrator"`
}

// workspaceAccessEntry mirrors the on-wire shape of one item in the
// /user/workspaces `values[]` array.
type workspaceAccessEntry struct {
	Administrator bool `json:"administrator"`
	Workspace     struct {
		UUID string `json:"uuid"`
		Slug string `json:"slug"`
	} `json:"workspace"`
}

// Atlassian's CHANGE-2770 removed every cross-workspace listing endpoint
// (/workspaces, /user/permissions/workspaces, /repositories) — they all
// return 410 Gone. The single replacement is /2.0/user/workspaces, a
// new endpoint released January 2026 specifically for this use case.
func workspaceList() error {
	c := apiClient()
	rows := make([]workspaceRow, 0)
	url := "/user/workspaces?pagelen=100"
	for page := 1; page <= 100; page++ {
		var resp struct {
			Values []workspaceAccessEntry `json:"values"`
			Next   string                 `json:"next"`
		}
		pageURL := url + "&page=" + strconv.Itoa(page)
		if err := c.JSON("GET", pageURL, nil, false, &resp); err != nil {
			return err
		}
		for _, e := range resp.Values {
			rows = append(rows, workspaceRow{
				Slug:          e.Workspace.Slug,
				UUID:          e.Workspace.UUID,
				Administrator: e.Administrator,
			})
		}
		if resp.Next == "" {
			break
		}
	}

	if render(rows) {
		return nil
	}
	for _, ws := range rows {
		util.O(map[string]string{
			"slug":          ws.Slug,
			"uuid":          ws.UUID,
			"administrator": yesNo(ws.Administrator),
		}, "yellow")
	}
	return nil
}

func init() {
	workspaceCmd.AddCommand(workspaceListCmd)
	rootCmd.AddCommand(workspaceCmd)
}
