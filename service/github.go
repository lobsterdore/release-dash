package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GithubService interface {
	GetChangelog(ctx context.Context, owner string, repo string, fromTag string, toTag string) (*github.CommitsComparison, error)
	GetDashboardReposFromOrg(ctx context.Context, org string) error
}

type githubService struct {
	Client *github.Client
}

func NewGithubService(ctx context.Context) GithubService {
	ghPat := os.Getenv("GH_PAT")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghPat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	service := githubService{
		Client: client,
	}

	return &service
}

func (c *githubService) GetChangelog(ctx context.Context, owner string, repo string, fromTag string, toTag string) (*github.CommitsComparison, error) {
	refFrom, _, err := c.Client.Git.GetRef(ctx, owner, repo, "tags/"+fromTag)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	refTo, _, err := c.Client.Git.GetRef(ctx, owner, repo, "tags/"+toTag)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	comparison, _, err := c.Client.Repositories.CompareCommits(ctx, owner, repo, *refFrom.Object.SHA, *refTo.Object.SHA)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return comparison, nil
}

func (c *githubService) GetDashboardReposFromOrg(ctx context.Context, org string) error {
	orgRepos, _, err := c.Client.Repositories.ListByOrg(ctx, org, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println(orgRepos)

	for _, repo := range orgRepos {
		fmt.Println(repo)
	}

	return nil
}
