package scm

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/flowchartsman/retry"
	"github.com/google/go-github/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type GithubAdapter struct {
	Client  *github.Client
	Retrier *retry.Retrier
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
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	return &service, nil
}

func CheckForRetry(resp *github.Response, err error) error {
	switch {
	case resp.StatusCode == 403:
		if _, ok := err.(*github.RateLimitError); ok {
			return fmt.Errorf("Retrying after rate limit response: %s", err)
		}
	case err != nil:
		return retry.Stop(err)

	}
	return nil
}

func (c *GithubAdapter) GetChangelog(ctx context.Context, owner string, repo string, fromTag string, toTag string) (*[]ScmCommit, error) {
	log.Debug().Msgf("Grabbing changelog for repo %s/%s, from-tag %s, to-tag %s", owner, repo, fromTag, toTag)

	var err error

	var refFrom *github.Reference
	var resp *github.Response
	err = c.Retrier.Run(func() error {
		var errReq error
		refFrom, resp, errReq = c.Client.Git.GetRef(ctx, owner, repo, "tags/"+fromTag)
		return CheckForRetry(resp, errReq)
	})
	if err != nil {
		if resp.StatusCode != 404 {
			return nil, fmt.Errorf("Could not get tag for repo: %s", err)
		}
		log.Debug().Msgf("Repo %s/%s does not have tag %s", owner, repo, fromTag)
	}

	var refTo *github.Reference
	err = c.Retrier.Run(func() error {
		var errReq error
		refTo, resp, errReq = c.Client.Git.GetRef(ctx, owner, repo, "tags/"+toTag)
		return CheckForRetry(resp, errReq)
	})
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

		var commits []*github.RepositoryCommit
		_ = c.Retrier.Run(func() error {
			var errReq error
			commits, resp, errReq = c.Client.Repositories.ListCommits(ctx, owner, repo, opt)
			return CheckForRetry(resp, errReq)
		})
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
	var repos []*github.Repository
	var resp *github.Response
	for {
		err := c.Retrier.Run(func() error {
			var errReq error
			repos, resp, errReq = c.Client.Repositories.List(ctx, user, opts)
			return CheckForRetry(resp, errReq)
		})
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
	var branches []*github.Branch
	var resp *github.Response
	err := c.Retrier.Run(func() error {
		var errReq error
		branches, resp, errReq = c.Client.Repositories.ListBranches(ctx, owner, repo, nil)
		return CheckForRetry(resp, errReq)
	})
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
	var repoTree *github.Tree
	var resp *github.Response
	err := c.Retrier.Run(func() error {
		var errReq error
		repoTree, resp, errReq = c.Client.Git.GetTree(ctx, owner, repo, sha, true)
		return CheckForRetry(resp, errReq)
	})
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
