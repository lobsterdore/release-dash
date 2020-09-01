package service

import (
	"context"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GithubService struct {
	Client *github.Client
}

func NewGithubService(ctx context.Context) *GithubService {
	ghPat := os.Getenv("GH_PAT")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghPat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	service := GithubService{
		Client: client,
	}

	return &service
}

func (c *GithubService) GetChangelog(ctx context.Context) (*github.CommitsComparison, error) {
	refFrom, _, err := c.Client.Git.GetRef(ctx, "lobsterdore", "lobstercms", "tags/v2.1.0")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	refTo, _, err := c.Client.Git.GetRef(ctx, "lobsterdore", "lobstercms", "tags/v2.7.0")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	comparison, _, err := c.Client.Repositories.CompareCommits(ctx, "lobsterdore", "lobstercms", *refFrom.Object.SHA, *refTo.Object.SHA)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return comparison, nil
}
