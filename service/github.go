package service

import (
	"context"
	"log"
	"os"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

func GetGithubClient(ctx context.Context) *github.Client {
	ghPat := os.Getenv("GH_PAT")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghPat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return client
}

func GetChangelog(ctx context.Context, client *github.Client) (*github.CommitsComparison, error) {
	refFrom, _, err := client.Git.GetRef(ctx, "lobsterdore", "lobstercms", "tags/v2.1.0")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	refTo, _, err := client.Git.GetRef(ctx, "lobsterdore", "lobstercms", "tags/v2.7.0")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	comparison, _, err := client.Repositories.CompareCommits(ctx, "lobsterdore", "lobstercms", *refFrom.Object.SHA, *refTo.Object.SHA)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return comparison, nil
}
