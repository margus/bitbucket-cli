package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/margus/bitbucket-cli/internal/util"
	"github.com/spf13/cobra"
)

var prDetailsCmd = &cobra.Command{
	Use:               "pr-details",
	Short:             "Show PR comments (general + inline). Same as `bb pr show`.",
	PersistentPreRunE: requireGit(),
}

var prDetailsShowCmd = &cobra.Command{
	Use:   "show <pr-id> [unresolved]",
	Short: "Show comments on a pull request",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		unresolved := len(args) == 2 && parseBoolish(args[1])
		return prDetailsShow(args[0], unresolved)
	},
}

type prComment struct {
	CreatedOn string `json:"created_on"`
	Deleted   bool   `json:"deleted"`
	User      struct {
		DisplayName string `json:"display_name"`
		UUID        string `json:"uuid"`
	} `json:"user"`
	Content struct {
		Raw string `json:"raw"`
	} `json:"content"`
	Inline *struct {
		Path string `json:"path"`
		To   *int   `json:"to"`
		From *int   `json:"from"`
	} `json:"inline,omitempty"`
	Resolution    json.RawMessage `json:"resolution,omitempty"`
	HasResolution bool            `json:"-"`
}

func prDetailsShow(prID string, unresolved bool) error {
	general, inline, err := fetchAllPrComments(prID, unresolved)
	if err != nil {
		return err
	}

	util.O(fmt.Sprintf("## Pull Request Comments (PR #%s)", prID), "green")
	util.O("", "white")

	util.O("### General Comments", "yellow")
	if len(general) == 0 {
		util.O("No general comments found.", "cyan")
	} else {
		for _, c := range general {
			renderComment(c, "")
		}
	}

	util.O("### Inline Code Comments", "yellow")
	if len(inline) == 0 {
		util.O("No inline comments found.", "cyan")
	} else {
		for _, c := range inline {
			line := ""
			if c.Inline != nil {
				if c.Inline.To != nil {
					line = fmt.Sprintf("%d", *c.Inline.To)
				} else if c.Inline.From != nil {
					line = fmt.Sprintf("%d", *c.Inline.From)
				}
			}
			path := ""
			if c.Inline != nil {
				path = c.Inline.Path
			}
			util.O(fmt.Sprintf("File: %s:%s", path, line), "cyan")
			renderComment(c, "")
		}
	}
	return nil
}

func renderComment(c prComment, _ string) {
	author := c.User.DisplayName
	if author == "" {
		author = "Unknown"
	}
	ts := util.FormatRelativeTimestamp(c.CreatedOn)
	util.O(fmt.Sprintf("%s (%s):", author, ts), "green")
	if c.Deleted {
		util.O("[DELETED]", "white")
	} else {
		util.O(c.Content.Raw, "white")
	}
	util.O("", "white")
}

func fetchAllPrComments(prID string, unresolved bool) ([]prComment, []prComment, error) {
	c := apiClient()
	const pagelen = 100
	var general, inline []prComment

	for page := 1; page <= 100; page++ {
		body, err := c.Request("GET", fmt.Sprintf("/pullrequests/%s/comments?pagelen=%d&page=%d", prID, pagelen, page), nil, true)
		if err != nil {
			return nil, nil, err
		}
		var raw struct {
			Values []map[string]json.RawMessage `json:"values"`
			Next   string                       `json:"next"`
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			return nil, nil, fmt.Errorf("decoding comments page: %w", err)
		}
		for _, m := range raw.Values {
			rebuilt, _ := json.Marshal(m)
			var pc prComment
			if err := json.Unmarshal(rebuilt, &pc); err != nil {
				continue
			}
			_, pc.HasResolution = m["resolution"]
			if pc.Inline != nil && pc.Inline.Path != "" {
				inline = append(inline, pc)
			} else {
				general = append(general, pc)
			}
		}
		if raw.Next == "" {
			break
		}
	}

	sort.SliceStable(general, func(i, j int) bool { return general[i].CreatedOn < general[j].CreatedOn })
	sort.SliceStable(inline, func(i, j int) bool { return inline[i].CreatedOn < inline[j].CreatedOn })

	if unresolved {
		filtered := inline[:0]
		for _, c := range inline {
			if !c.HasResolution {
				filtered = append(filtered, c)
			}
		}
		inline = filtered
	}
	return general, inline, nil
}

func init() {
	prDetailsCmd.AddCommand(prDetailsShowCmd)
	rootCmd.AddCommand(prDetailsCmd)
}
