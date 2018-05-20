package main

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	// Period is used for selecting the range of Gist contributions.
	Period string

	// GistUser is used for retrieving the user's Gist.
	GistUser string
)

// gistPages wraps the original Gist to define the method
type gistPages []*github.Gist

// sortByUpdatedAt sorts the Gist pages by updated time in descending order.
func (pages gistPages) sortByUpdatedAt() gistPages {
	sort.SliceStable(pages, func(i, j int) bool {
		return pages[i].GetUpdatedAt().After(pages[j].GetUpdatedAt())
	})
	return pages
}

// stringifyGistHistory returns the formatted description of Gist history.
func (pages gistPages) stringifyGistHistory() string {
	var buf bytes.Buffer
	buf.WriteString(":octocat: *Gist Activities* :octocat:\n")
	for _, page := range pages {
		buf.WriteString("[")
		buf.WriteString(page.GetUpdatedAt().Format("2006-01-02"))
		buf.WriteString("]: <")
		buf.WriteString(page.GetGitPushURL())
		buf.WriteString("|")
		buf.WriteString(page.GetDescription())
		buf.WriteString(">\n")
	}
	return buf.String()
}

// setAuthToken sets the Github personal token if specified.
func setAuthToken(client *github.Client, token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// getGistHistory shows the contributions activity during the specified period.
func getGistHistory(h interactionHandler, period string) (string, error) {
	var since time.Time
	switch period {
	case "weekly":
		since = time.Now().AddDate(0, 0, -7)
	case "monthly":
		since = time.Now().AddDate(0, -1, 0)
	case "yearly":
		since = time.Now().AddDate(-1, 0, 0)
	}

	client := github.NewClient(nil)
	if h.gistAccessToken != "" {
		client = setAuthToken(client, h.gistAccessToken)
	}

	var gists gistPages
	var err error
	gists, _, err = client.Gists.List(context.Background(), GistUser,
		&github.GistListOptions{Since: since},
	)
	if err != nil {
		return "", fmt.Errorf("Failed to get list of Gist contributions: %s", err)
	}

	return gists.sortByUpdatedAt().stringifyGistHistory(), nil
}
