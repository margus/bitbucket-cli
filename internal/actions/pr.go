package actions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/margus/bitbucket-cli/internal/api"
	"github.com/margus/bitbucket-cli/internal/util"
)

// PRTitle and PRDescription are populated from --title/--description flags
// in main before Dispatch is called.
var (
	PRTitle       string
	PRDescription string
	Interactive   bool
)

type Pr struct{}

func (Pr) Default() string { return "list" }
func (Pr) Commands() map[string]string {
	return map[string]string{
		"list":             "list, l",
		"diff":             "diff, d",
		"files":            "files",
		"commits":          "commits, c",
		"approve":          "approve, a",
		"unApprove":        "no-approve, na",
		"requestChanges":   "request-changes, rc",
		"unRequestChanges": "no-request-changes, nrc",
		"decline":          "decline",
		"merge":            "merge, m",
		"create":           "create",
		"show":             "show",
	}
}
func (Pr) RequireGit() bool { return true }

func (p Pr) Dispatch(method string, args []string) error {
	switch method {
	case "list":
		dst := ""
		if len(args) > 0 {
			dst = args[0]
		}
		return p.list(dst)
	case "diff":
		if len(args) < 1 {
			return fmt.Errorf("diff requires <pr-number>")
		}
		return p.diff(args[0])
	case "files":
		if len(args) < 1 {
			return fmt.Errorf("files requires <pr-number>")
		}
		return p.files(args[0])
	case "commits":
		if len(args) < 1 {
			return fmt.Errorf("commits requires <pr-number>")
		}
		return p.commits(args[0])
	case "approve":
		if len(args) < 1 {
			return fmt.Errorf("pr number required")
		}
		return p.approve(args)
	case "unApprove":
		if len(args) < 1 {
			return fmt.Errorf("unApprove requires <pr-number>")
		}
		return p.simpleDelete(args[0], "approve")
	case "requestChanges":
		if len(args) < 1 {
			return fmt.Errorf("requestChanges requires <pr-number>")
		}
		return p.simplePost(args[0], "request-changes")
	case "unRequestChanges":
		if len(args) < 1 {
			return fmt.Errorf("unRequestChanges requires <pr-number>")
		}
		return p.simpleDelete(args[0], "request-changes")
	case "decline":
		if len(args) < 1 {
			return fmt.Errorf("decline requires <pr-number>")
		}
		return p.decline(args[0])
	case "merge":
		if len(args) < 1 {
			return fmt.Errorf("merge requires <pr-number>")
		}
		return p.merge(args[0])
	case "create":
		if len(args) < 1 {
			return fmt.Errorf("create requires at least <to-branch>")
		}
		fromBranch := ""
		toBranch := ""
		addReviewers := 1
		if len(args) == 1 {
			toBranch = args[0]
		} else {
			fromBranch = args[0]
			toBranch = args[1]
		}
		if len(args) >= 3 {
			if v, err := strconv.Atoi(args[2]); err == nil {
				addReviewers = v
			}
		}
		return p.create(fromBranch, toBranch, addReviewers)
	case "show":
		if len(args) < 1 {
			return fmt.Errorf("show requires <pr-id>")
		}
		unresolved := false
		if len(args) >= 2 {
			unresolved = parseBool(args[1])
		}
		return PrDetails{}.show(args[0], unresolved)
	default:
		return fmt.Errorf("unknown method: %s", method)
	}
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

func (Pr) list(destination string) error {
	c := api.New()
	var list struct {
		Values []prListItem `json:"values"`
	}
	if err := c.JSON("GET", "/pullrequests?state=OPEN", nil, true, &list); err != nil {
		return err
	}
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
		util.O(map[string]string{
			"id":           strconv.Itoa(pr.ID),
			"author":       pr.Author.Nickname,
			"source":       pr.Source.Branch.Name,
			"destination":  pr.Destination.Branch.Name,
			"link":         pr.Links.HTML.Href,
			"reviewers":    strings.Join(reviewers, ", "),
			"participants": strings.Join(participants, " | "),
		}, "yellow")
	}
	return nil
}

func (Pr) diff(prNumber string) error {
	c := api.New()
	b, err := c.Request("GET", "/pullrequests/"+prNumber+"/diff", nil, true)
	if err != nil {
		return err
	}
	util.O(string(b), "yellow")
	return nil
}

func (Pr) files(prNumber string) error {
	c := api.New()
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

func (Pr) commits(prNumber string) error {
	c := api.New()
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

func (p Pr) approve(prNumbers []string) error {
	if prNumbers[0] == "0" {
		// Approve all open.
		c := api.New()
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
			return fmt.Errorf("pr not found")
		}
	}
	c := api.New()
	for _, n := range prNumbers {
		if _, err := c.Request("POST", "/pullrequests/"+n+"/approve", nil, true); err != nil {
			return err
		}
		util.O(n+" Approved.", "green")
	}
	return nil
}

func (Pr) simplePost(prNumber, suffix string) error {
	c := api.New()
	b, err := c.Request("POST", "/pullrequests/"+prNumber+"/"+suffix, nil, true)
	if err != nil {
		return err
	}
	printRaw(b)
	return nil
}

func (Pr) simpleDelete(prNumber, suffix string) error {
	c := api.New()
	b, err := c.Request("DELETE", "/pullrequests/"+prNumber+"/"+suffix, nil, true)
	if err != nil {
		return err
	}
	printRaw(b)
	return nil
}

func (Pr) decline(prNumber string) error {
	c := api.New()
	if _, err := c.Request("POST", "/pullrequests/"+prNumber+"/decline", nil, true); err != nil {
		return err
	}
	util.O("OK.", "green")
	return nil
}

func (Pr) merge(prNumber string) error {
	c := api.New()
	var resp struct {
		State string `json:"state"`
	}
	if err := c.JSON("POST", "/pullrequests/"+prNumber+"/merge", nil, true, &resp); err != nil {
		return err
	}
	util.O(resp.State, "green")
	return nil
}

func (p Pr) create(fromBranch, toBranch string, addDefaultReviewers int) error {
	if fromBranch == "" {
		// Single-arg form: arg is the to-branch; derive from-branch from HEAD.
		current, err := util.CurrentBranch()
		if err != nil {
			return err
		}
		fromBranch = current
	}

	title := PRTitle
	description := PRDescription
	if Interactive {
		if title == "" {
			title = util.Prompt("PR title (leave empty for default):", "")
		}
		if description == "" {
			description = util.Prompt("PR description (leave empty to skip):", "")
		}
	}

	return p.bulkCreate(strings.Split(toBranch, ","), fromBranch, addDefaultReviewers == 1, title, description)
}

func (p Pr) bulkCreate(toBranches []string, fromBranch string, addDefaultReviewers bool, title, description string) error {
	c := api.New()
	var reviewers []map[string]any
	if addDefaultReviewers {
		var err error
		reviewers, err = p.defaultReviewers(c)
		if err != nil {
			return err
		}
	}

	type result struct {
		ID   int    `json:"id"`
		Link string `json:"link"`
	}
	var responses []result

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
		responses = append(responses, result{ID: resp.ID, Link: resp.Links.HTML.Href})
	}

	for _, r := range responses {
		util.O(map[string]string{
			"id":   strconv.Itoa(r.ID),
			"link": r.Link,
		}, "yellow")
	}
	return nil
}

func (Pr) defaultReviewers(c *api.Client) ([]map[string]any, error) {
	// current user uuid
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

func printRaw(b []byte) {
	// Pretty-print JSON when possible, otherwise dump string.
	var v any
	if err := json.Unmarshal(b, &v); err == nil {
		util.O(v, "yellow")
		return
	}
	if len(b) > 0 {
		util.O(string(b), "yellow")
	}
}
