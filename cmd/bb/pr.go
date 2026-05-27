package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/margus/bitbucket-cli/internal/api"
	"github.com/margus/bitbucket-cli/internal/util"
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:               "pr",
	Short:             "Bitbucket pull requests: list, view, approve, merge, etc.",
	PersistentPreRunE: requireGit(),
	RunE: func(cmd *cobra.Command, args []string) error {
		dst := ""
		if len(args) > 0 {
			dst = args[0]
		}
		return prList(dst)
	},
}

var prListCmd = &cobra.Command{
	Use:     "list [destination-branch]",
	Aliases: []string{"l"},
	Short:   "List open pull requests (optionally filtered by destination branch)",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dst := ""
		if len(args) > 0 {
			dst = args[0]
		}
		return prList(dst)
	},
}

var prDiffCmd = &cobra.Command{
	Use:     "diff <pr-number>",
	Aliases: []string{"d"},
	Short:   "Print the unified diff for a pull request",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prDiff(args[0])
	},
}

var prFilesCmd = &cobra.Command{
	Use:   "files <pr-number>",
	Short: "List files changed in a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prFiles(args[0])
	},
}

var prCommitsCmd = &cobra.Command{
	Use:     "commits <pr-number>",
	Aliases: []string{"c"},
	Short:   "List commit messages in a pull request",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prCommits(args[0])
	},
}

var prApproveCmd = &cobra.Command{
	Use:     "approve <pr-number>...",
	Aliases: []string{"a"},
	Short:   "Approve one or more pull requests (use `approve 0` to approve all open)",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prApprove(args)
	},
}

var prUnApproveCmd = &cobra.Command{
	Use:     "no-approve <pr-number>",
	Aliases: []string{"na"},
	Short:   "Revert an approval",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prSimpleDelete(args[0], "approve")
	},
}

var prRequestChangesCmd = &cobra.Command{
	Use:     "request-changes <pr-number>",
	Aliases: []string{"rc"},
	Short:   "Request changes on a pull request",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prSimplePost(args[0], "request-changes")
	},
}

var prUnRequestChangesCmd = &cobra.Command{
	Use:     "no-request-changes <pr-number>",
	Aliases: []string{"nrc"},
	Short:   "Revert a change-request review",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prSimpleDelete(args[0], "request-changes")
	},
}

var prDeclineCmd = &cobra.Command{
	Use:   "decline <pr-number>",
	Short: "Decline a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prDecline(args[0])
	},
}

var prMergeCmd = &cobra.Command{
	Use:     "merge <pr-number>",
	Aliases: []string{"m"},
	Short:   "Merge a pull request",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prMerge(args[0])
	},
}

var prCreateCmd = &cobra.Command{
	Use:   "create <from-branch> <to-branch> [add-default-reviewers]",
	Short: "Create a pull request",
	Long: `Create a pull request from <from-branch> to <to-branch>. Pass <to-branch>
only to use the current HEAD as the source. Pass 0 as the third arg to
skip default reviewers.

Use --title/--description on the root command to set them non-interactively.
Use -i/--interactive to be prompted.`,
	Args: cobra.RangeArgs(1, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		from, to := "", ""
		addReviewers := 1
		if len(args) == 1 {
			to = args[0]
		} else {
			from = args[0]
			to = args[1]
		}
		if len(args) >= 3 {
			if v, err := strconv.Atoi(args[2]); err == nil {
				addReviewers = v
			}
		}
		return prCreate(from, to, addReviewers)
	},
}

var prCheckoutCmd = &cobra.Command{
	Use:   "checkout <pr-id>",
	Short: "Fetch and check out the source branch of a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prCheckout(args[0])
	},
}

var prViewCmd = &cobra.Command{
	Use:   "view <pr-id>",
	Short: "Open a pull request in the default browser",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prView(args[0])
	},
}

var prCommentCmd = &cobra.Command{
	Use:   "comment <pr-id> <message>",
	Short: "Post a general comment on a pull request",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return prAddComment(args[0], args[1])
	},
}

var prCommentInlineCmd = &cobra.Command{
	Use:   "comment-inline <pr-id> <file> <line> <message>",
	Short: "Post an inline review comment on a specific file and line",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		line, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("line must be an integer: %q", args[2])
		}
		return prAddCommentInline(args[0], args[1], line, args[3])
	},
}

var prShowCmd = &cobra.Command{
	Use:   "show <pr-id> [unresolved]",
	Short: "Show comments on a pull request. Pass `true` to show only unresolved inline comments.",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		unresolved := len(args) == 2 && parseBoolish(args[1])
		return prDetailsShow(args[0], unresolved)
	},
}

type prListItem struct {
	ID     int `json:"id"`
	Author struct {
		Nickname string `json:"nickname"`
	} `json:"author"`
	Source struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"source"`
	Destination struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"destination"`
	Links struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

type prDetail struct {
	Reviewers []struct {
		DisplayName string `json:"display_name"`
	} `json:"reviewers"`
	Participants []struct {
		State string `json:"state"`
		User  struct {
			DisplayName string `json:"display_name"`
		} `json:"user"`
	} `json:"participants"`
}

type prListRow struct {
	ID           int    `json:"id" yaml:"id"`
	Author       string `json:"author" yaml:"author"`
	Source       string `json:"source" yaml:"source"`
	Destination  string `json:"destination" yaml:"destination"`
	Link         string `json:"link" yaml:"link"`
	Reviewers    string `json:"reviewers" yaml:"reviewers"`
	Participants string `json:"participants" yaml:"participants"`
}

func prList(destination string) error {
	c := apiClient()
	var list struct {
		Values []prListItem `json:"values"`
	}
	if err := c.JSON("GET", "/pullrequests?state=OPEN", nil, true, &list); err != nil {
		return err
	}
	rows := make([]prListRow, 0, len(list.Values))
	for _, pr := range list.Values {
		if destination != "" && pr.Destination.Branch.Name != destination {
			continue
		}
		var d prDetail
		if err := c.JSON("GET", fmt.Sprintf("/pullrequests/%d", pr.ID), nil, true, &d); err != nil {
			return err
		}
		reviewers := make([]string, 0, len(d.Reviewers))
		for _, r := range d.Reviewers {
			reviewers = append(reviewers, r.DisplayName)
		}
		participants := make([]string, 0, len(d.Participants))
		for _, p := range d.Participants {
			if p.State == "" {
				continue
			}
			participants = append(participants, fmt.Sprintf("%s -> %s", p.User.DisplayName, p.State))
		}
		rows = append(rows, prListRow{
			ID:           pr.ID,
			Author:       pr.Author.Nickname,
			Source:       pr.Source.Branch.Name,
			Destination:  pr.Destination.Branch.Name,
			Link:         pr.Links.HTML.Href,
			Reviewers:    strings.Join(reviewers, ", "),
			Participants: strings.Join(participants, " | "),
		})
	}
	if render(rows) {
		return nil
	}
	for _, r := range rows {
		util.O(map[string]string{
			"id":           strconv.Itoa(r.ID),
			"author":       r.Author,
			"source":       r.Source,
			"destination":  r.Destination,
			"link":         r.Link,
			"reviewers":    r.Reviewers,
			"participants": r.Participants,
		}, "yellow")
	}
	return nil
}

func prDiff(prNumber string) error {
	c := apiClient()
	b, err := c.Request("GET", "/pullrequests/"+prNumber+"/diff", nil, true)
	if err != nil {
		return err
	}
	util.O(string(b), "yellow")
	return nil
}

func prFiles(prNumber string) error {
	c := apiClient()
	var resp struct {
		Values []struct {
			New struct {
				Path string `json:"path"`
			} `json:"new"`
		} `json:"values"`
	}
	if err := c.JSON("GET", "/pullrequests/"+prNumber+"/diffstat", nil, true, &resp); err != nil {
		return err
	}
	for _, row := range resp.Values {
		util.O(row.New.Path, "yellow")
	}
	return nil
}

func prCommits(prNumber string) error {
	c := apiClient()
	var resp struct {
		Values []struct {
			Summary struct {
				Raw string `json:"raw"`
			} `json:"summary"`
		} `json:"values"`
	}
	if err := c.JSON("GET", "/pullrequests/"+prNumber+"/commits", nil, true, &resp); err != nil {
		return err
	}
	for _, v := range resp.Values {
		util.O(strings.TrimSpace(strings.ReplaceAll(v.Summary.Raw, `\n`, "\n")), "yellow")
	}
	return nil
}

func prApprove(prNumbers []string) error {
	if prNumbers[0] == "0" {
		c := apiClient()
		var list struct {
			Values []prListItem `json:"values"`
		}
		if err := c.JSON("GET", "/pullrequests?state=OPEN", nil, true, &list); err != nil {
			return err
		}
		prNumbers = prNumbers[:0]
		for _, pr := range list.Values {
			prNumbers = append(prNumbers, strconv.Itoa(pr.ID))
		}
		if len(prNumbers) == 0 {
			return fmt.Errorf("no open PRs found")
		}
	}
	c := apiClient()
	for _, n := range prNumbers {
		if _, err := c.Request("POST", "/pullrequests/"+n+"/approve", nil, true); err != nil {
			return err
		}
		util.O(n+" Approved.", "green")
	}
	return nil
}

func prSimplePost(prNumber, suffix string) error {
	c := apiClient()
	b, err := c.Request("POST", "/pullrequests/"+prNumber+"/"+suffix, nil, true)
	if err != nil {
		return err
	}
	prPrintRaw(b)
	return nil
}

func prSimpleDelete(prNumber, suffix string) error {
	c := apiClient()
	b, err := c.Request("DELETE", "/pullrequests/"+prNumber+"/"+suffix, nil, true)
	if err != nil {
		return err
	}
	prPrintRaw(b)
	return nil
}

func prDecline(prNumber string) error {
	c := apiClient()
	if _, err := c.Request("POST", "/pullrequests/"+prNumber+"/decline", nil, true); err != nil {
		return err
	}
	util.O("OK.", "green")
	return nil
}

// prCheckoutGitExec is the git command runner. Swappable for tests.
var prCheckoutGitExec = func(args ...string) ([]byte, error) {
	return exec.Command("git", args...).CombinedOutput()
}

func prCheckout(prNumber string) error {
	c := apiClient()
	var resp struct {
		Source struct {
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"source"`
	}
	if err := c.JSON("GET", "/pullrequests/"+prNumber, nil, true, &resp); err != nil {
		return err
	}
	branch := resp.Source.Branch.Name
	if branch == "" {
		return fmt.Errorf("could not determine source branch of PR %s", prNumber)
	}
	util.O(fmt.Sprintf("Fetching %s ...", branch), "yellow")
	if out, err := prCheckoutGitExec("fetch", "origin", branch); err != nil {
		return fmt.Errorf("git fetch: %w\n%s", err, out)
	}
	if out, err := prCheckoutGitExec("checkout", branch); err != nil {
		return fmt.Errorf("git checkout: %w\n%s", err, out)
	}
	util.O("Checked out "+branch, "green")
	return nil
}

// prViewOpener is the browser launcher. Swappable for tests.
var prViewOpener = func(url string) error {
	opener := openerFor(runtime.GOOS)
	if opener == "" {
		return fmt.Errorf("don't know how to open a browser on %q", runtime.GOOS)
	}
	return exec.Command(opener, url).Start()
}

func prView(prNumber string) error {
	c := apiClient()
	var resp struct {
		Links struct {
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
		} `json:"links"`
	}
	if err := c.JSON("GET", "/pullrequests/"+prNumber, nil, true, &resp); err != nil {
		return err
	}
	url := resp.Links.HTML.Href
	if url == "" {
		return fmt.Errorf("no html link returned for PR %s", prNumber)
	}
	util.O(url, "white")
	return prViewOpener(url)
}

func prAddComment(prNumber, message string) error {
	c := apiClient()
	payload := map[string]any{
		"content": map[string]any{"raw": message},
	}
	var resp struct {
		ID   int `json:"id"`
		User struct {
			DisplayName string `json:"display_name"`
		} `json:"user"`
	}
	if err := c.JSON("POST", "/pullrequests/"+prNumber+"/comments", payload, true, &resp); err != nil {
		return err
	}
	util.O(map[string]string{
		"id":     strconv.Itoa(resp.ID),
		"author": resp.User.DisplayName,
	}, "yellow")
	return nil
}

func prAddCommentInline(prNumber, file string, line int, message string) error {
	c := apiClient()
	payload := map[string]any{
		"content": map[string]any{"raw": message},
		"inline": map[string]any{
			"path": file,
			"to":   line, // line in the new file; use `from` for old-side comments
		},
	}
	var resp struct {
		ID     int `json:"id"`
		Inline struct {
			Path string `json:"path"`
			To   *int   `json:"to"`
		} `json:"inline"`
	}
	if err := c.JSON("POST", "/pullrequests/"+prNumber+"/comments", payload, true, &resp); err != nil {
		return err
	}
	util.O(map[string]string{
		"id":   strconv.Itoa(resp.ID),
		"file": resp.Inline.Path,
		"line": strconv.Itoa(line),
	}, "yellow")
	return nil
}

func prMerge(prNumber string) error {
	c := apiClient()
	var resp struct {
		State string `json:"state"`
	}
	if err := c.JSON("POST", "/pullrequests/"+prNumber+"/merge", nil, true, &resp); err != nil {
		return err
	}
	util.O(resp.State, "green")
	return nil
}

func prCreate(fromBranch, toBranch string, addDefaultReviewers int) error {
	if fromBranch == "" {
		current, err := util.CurrentBranch()
		if err != nil {
			return err
		}
		fromBranch = current
	}

	title := flagTitle
	description := flagDescription
	if flagInteractive {
		if title == "" {
			title = util.Prompt("PR title (leave empty for default):", "")
		}
		if description == "" {
			description = util.Prompt("PR description (leave empty to skip):", "")
		}
	}
	return prBulkCreate(strings.Split(toBranch, ","), fromBranch, addDefaultReviewers == 1, title, description)
}

func prBulkCreate(toBranches []string, fromBranch string, addDefaultReviewers bool, title, description string) error {
	c := apiClient()
	var reviewers []map[string]any
	if addDefaultReviewers {
		var err error
		reviewers, err = prDefaultReviewers(c)
		if err != nil {
			return err
		}
	}

	for _, to := range toBranches {
		actualTitle := title
		if actualTitle == "" {
			actualTitle = fmt.Sprintf("Merge %s into %s", fromBranch, to)
		}
		payload := map[string]any{
			"title": actualTitle,
			"source": map[string]any{
				"branch": map[string]any{"name": fromBranch},
			},
			"destination": map[string]any{
				"branch": map[string]any{"name": to},
			},
			"reviewers": reviewers,
		}
		if description != "" {
			payload["description"] = description
		}
		var resp struct {
			ID    int `json:"id"`
			Links struct {
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
			} `json:"links"`
		}
		if err := c.JSON("POST", "/pullrequests", payload, true, &resp); err != nil {
			return err
		}
		util.O(map[string]string{
			"id":   strconv.Itoa(resp.ID),
			"link": resp.Links.HTML.Href,
		}, "yellow")
	}
	return nil
}

func prDefaultReviewers(c *api.Client) ([]map[string]any, error) {
	var me struct {
		UUID string `json:"uuid"`
	}
	if err := c.JSON("GET", "/user", nil, false, &me); err != nil {
		return nil, err
	}
	var resp struct {
		Values []map[string]any `json:"values"`
	}
	if err := c.JSON("GET", "/default-reviewers", nil, true, &resp); err != nil {
		return nil, err
	}
	out := resp.Values[:0]
	for _, r := range resp.Values {
		if u, _ := r["uuid"].(string); u != me.UUID {
			out = append(out, r)
		}
	}
	return out, nil
}

func prPrintRaw(b []byte) {
	var v any
	if err := json.Unmarshal(b, &v); err == nil {
		util.O(v, "yellow")
		return
	}
	if len(b) > 0 {
		util.O(string(b), "yellow")
	}
}

// parseBoolish accepts "true"/"false"/"1"/"0" and returns false for
// anything else — kept tolerant for legacy CLI usage like `bb pr show
// 123 unresolved`.
func parseBoolish(s string) bool {
	switch strings.ToLower(s) {
	case "true", "1", "yes", "y", "unresolved":
		return true
	}
	return false
}

func init() {
	prCmd.AddCommand(
		prListCmd, prDiffCmd, prFilesCmd, prCommitsCmd,
		prApproveCmd, prUnApproveCmd, prRequestChangesCmd, prUnRequestChangesCmd,
		prDeclineCmd, prMergeCmd, prCreateCmd, prShowCmd,
		prCommentCmd, prCommentInlineCmd,
		prCheckoutCmd, prViewCmd,
	)
	rootCmd.AddCommand(prCmd)
}
