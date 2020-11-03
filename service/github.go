package service

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=-mod=mod --source=github.go --destination=../mocks/service/github.go
type GithubProvider interface {
	GetChangelog(ctx context.Context, owner string, repo string, fromTag string, toTag string) (*github.CommitsComparison, error)
	GetRepoBranch(ctx context.Context, owner string, repo string, branchName string) (*github.Branch, error)
	GetRepoFile(ctx context.Context, owner string, repo string, sha string, filePath string) (*github.RepositoryContent, error)
	GetUserRepos(ctx context.Context, user string) ([]*github.Repository, error)
}

type GithubService struct {
	Client *github.Client
}

func NewGithubService(ctx context.Context, pat string) GithubProvider {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	service := GithubService{
		Client: client,
	}

	return &service
}

func (c *GithubService) GetChangelog(ctx context.Context, owner string, repo string, fromTag string, toTag string) (*github.CommitsComparison, error) {
	refFrom, _, err := c.Client.Git.GetRef(ctx, owner, repo, "tags/"+fromTag)
	if err != nil {
		log.Error().Err(err).Msg("Could not get repo from tag")
		return nil, err
	}

	refTo, _, err := c.Client.Git.GetRef(ctx, owner, repo, "tags/"+toTag)
	if err != nil {
		log.Error().Err(err).Msg("Could not get repo to tag")
		return nil, err
	}

	comparison, _, err := c.Client.Repositories.CompareCommits(ctx, owner, repo, *refFrom.Object.SHA, *refTo.Object.SHA)
	if err != nil {
		log.Error().Err(err).Msg("Could not get repo tag compare")
		return nil, err
	}

	return comparison, nil
}

func (c *GithubService) GetUserRepos(ctx context.Context, user string) ([]*github.Repository, error) {
	opts := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepos []*github.Repository
	for {
		repos, resp, err := c.Client.Repositories.List(ctx, "", opts)
		if err != nil {
			log.Error().Err(err).Msg("Could not get user repos")
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allRepos, nil
}

func (c *GithubService) GetRepoBranch(ctx context.Context, owner string, repo string, branchName string) (*github.Branch, error) {
	branches, _, err := c.Client.Repositories.ListBranches(ctx, owner, repo, nil)
	if err != nil {
		log.Error().Err(err).Msg("Could not get repo branches")
		return nil, err
	}

	for _, branch := range branches {
		if *branch.Name == branchName {
			return branch, nil
		}
	}

	return nil, nil
}

func (c *GithubService) GetRepoFile(ctx context.Context, owner string, repo string, sha string, filePath string) (*github.RepositoryContent, error) {
	repoTree, _, err := c.Client.Git.GetTree(ctx, owner, repo, sha, true)
	if err != nil {
		log.Error().Err(err).Msg("Could not get repo tree")
		return nil, err
	}

	for _, treeEntry := range repoTree.Entries {
		if *treeEntry.Path == filePath {
			content, _, _, err := c.Client.Repositories.GetContents(ctx, owner, repo, filePath, nil)
			if err != nil {
				log.Error().Err(err).Msg("Could not get repo file contents")
				return nil, err
			}
			return content, nil
		}
	}

	return nil, nil
}
