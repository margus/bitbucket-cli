package actions

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/margus/bitbucket-cli/internal/api"
	"github.com/margus/bitbucket-cli/internal/util"
)

type PrDetails struct{}

func (PrDetails) Default() string                { return "show" }
func (PrDetails) Commands() map[string]string    { return map[string]string{"show": "show"} }
func (PrDetails) RequireGit() bool               { return true }

func (p PrDetails) Dispatch(method string, args []string) error {
	switch method {
	case "show":
		if len(args) < 1 {
			return fmt.Errorf("PR ID required. Usage: bb pr-details show <pr_id> [unresolved]")
		}
		unresolved := false
		if len(args) >= 2 {
			unresolved = parseBool(args[1])
		}
		return p.show(args[0], unresolved)
	default:
		return fmt.Errorf("unknown method: %s", method)
	}
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
	Resolution json.RawMessage `json:"resolution,omitempty"`
	// HasResolution indicates whether the field was present in the JSON
	// (Bitbucket returns the key for resolved comments — even {} counts).
	HasResolution bool `json:"-"`
}

func (p PrDetails) show(prID string, unresolved bool) error {
	general, inline, err := p.fetchAllComments(prID, unresolved)
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
			author := c.User.DisplayName
			if author == "" {
				author = "Unknown"
			}
			ts := util.FormatRelativeTimestamp(c.CreatedOn)
			util.O(fmt.Sprintf("File: %s:%s", path, line), "cyan")
			util.O(fmt.Sprintf("%s (%s):", author, ts), "green")
			if c.Deleted {
				util.O("[DELETED]", "white")
			} else {
				util.O(c.Content.Raw, "white")
			}
			util.O("", "white")
		}
	}
	return nil
}

func (PrDetails) fetchAllComments(prID string, unresolved bool) ([]prComment, []prComment, error) {
	c := api.New()
	const pagelen = 100
	var general, inline []prComment

	for page := 1; page <= 100; page++ {
		body, err := c.Request("GET", fmt.Sprintf("/pullrequests/%s/comments?pagelen=%d&page=%d", prID, pagelen, page), nil, true)
		if err != nil {
			return nil, nil, err
		}
		// Decode in two passes: first raw, so we can detect whether
		// "resolution" appeared as a key — Bitbucket sets that field
		// (even as {}) on resolved comments.
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
