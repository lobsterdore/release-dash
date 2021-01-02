package scm

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/google/go-github/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type GithubAdapter struct {
	Client *github.Client
}

func NewGithubAdapter(ctx context.Context, pat string, urlDefault string, urlUpload string) (*GithubAdapter, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	if urlDefault != "" && urlUpload != "" {
		parsedUrlDefault, err := url.Parse(urlDefault)
		if err != nil {
			return nil, err
		}
		parsedUrlUpload, err := url.Parse(urlUpload)
		if err != nil {
			return nil, err
		}
		client.BaseURL = parsedUrlDefault
		client.UploadURL = parsedUrlUpload
	}

	service := GithubAdapter{
		Client: client,
	}

	return &service, nil
}

func (c *GithubAdapter) GetChangelog(ctx context.Context, owner string, repo string, fromTag string, toTag string) (*[]ScmCommit, error) {
	log.Debug().Msgf("Grabbing changelog for repo %s/%s, from-tag %s, to-tag %s", owner, repo, fromTag, toTag)

	refFrom, resp, err := c.Client.Git.GetRef(ctx, owner, repo, "tags/"+fromTag)
	if err != nil {
		if resp.StatusCode != 404 {
			return nil, fmt.Errorf("Could not get tag for repo: %s", err)
		}
		log.Debug().Msgf("Repo %s/%s does not have tag %s", owner, repo, fromTag)
	}

	refTo, resp, err := c.Client.Git.GetRef(ctx, owner, repo, "tags/"+toTag)
	if err != nil {
		if resp.StatusCode != 404 {
			return nil, fmt.Errorf("Could not get tag for repo: %s", err)
		}
		log.Debug().Msgf("Repo %s/%s does not have tag %s", owner, repo, toTag)
	}

	if refTo == nil {
		return nil, nil
	}

	comparison := &github.CommitsComparison{}
	if refFrom == nil {
		opt := &github.CommitsListOptions{
			SHA: toTag,
		}
		commits, _, _ := c.Client.Repositories.ListCommits(ctx, owner, repo, opt)
		if len(commits) == 0 {
			log.Debug().Msgf("Repo %s/%s does not have any commits", owner, repo)
			return nil, nil
		}
		comparison, _, err = c.Client.Repositories.CompareCommits(ctx, owner, repo, *commits[len(commits)-1].SHA, *refTo.Object.SHA)
		if err != nil {
			return nil, fmt.Errorf("Could not get repo tag comparison: %s", err)
		}
	} else {
		comparison, _, err = c.Client.Repositories.CompareCommits(ctx, owner, repo, *refFrom.Object.SHA, *refTo.Object.SHA)
		if err != nil {
			return nil, fmt.Errorf("Could not get repo tag comparison: %s", err)
		}
	}

	var allScmCommits []ScmCommit
	for _, commit := range comparison.Commits {
		scmCommit := ScmCommit{
			AuthorAvatarUrl: *commit.Author.AvatarURL,
			Message:         *commit.Commit.Message,
			HtmlUrl:         *commit.HTMLURL,
		}
		allScmCommits = append(allScmCommits, scmCommit)
	}

	return &allScmCommits, nil
}

func (c *GithubAdapter) GetUserRepos(ctx context.Context, user string) ([]ScmRepository, error) {
	opts := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepos []*github.Repository
	for {
		repos, resp, err := c.Client.Repositories.List(ctx, user, opts)
		if err != nil {
			return nil, fmt.Errorf("Could not get user repos: %s", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	var allScmRepos []ScmRepository

	for _, repo := range allRepos {
		scmRepo := ScmRepository{
			DefaultBranch: *repo.DefaultBranch,
			HtmlUrl:       *repo.HTMLURL,
			Name:          *repo.Name,
			OwnerName:     *repo.Owner.Login,
		}
		allScmRepos = append(allScmRepos, scmRepo)
	}

	return allScmRepos, nil
}

func (c *GithubAdapter) GetRepoBranch(ctx context.Context, owner string, repo string, branchName string) (*ScmBranch, error) {
	branches, _, err := c.Client.Repositories.ListBranches(ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not get repo branches: %s", err)
	}

	for _, branch := range branches {
		if *branch.Name == branchName {
			scmBranch := ScmBranch{
				CurrentHash: *branch.Commit.SHA,
				Name:        *branch.Name,
			}
			return &scmBranch, nil
		}
	}

	return nil, nil
}

func (c *GithubAdapter) GetRepoFile(ctx context.Context, owner string, repo string, sha string, filePath string) ([]byte, error) {
	repoTree, _, err := c.Client.Git.GetTree(ctx, owner, repo, sha, true)
	if err != nil {
		return nil, fmt.Errorf("Could not get repo tree: %s", err)
	}

	for _, treeEntry := range repoTree.Entries {
		if *treeEntry.Path == filePath {
			content, _, _, err := c.Client.Repositories.GetContents(ctx, owner, repo, filePath, nil)
			if err != nil {
				return nil, fmt.Errorf("Could not get repo file contents: %s", err)
			}

			raw, err := base64.StdEncoding.DecodeString(*content.Content)
			if err != nil {
				return nil, fmt.Errorf("Could not decode repo file: %s", err)
			}

			return raw, nil
		}
	}

	return nil, nil
}
