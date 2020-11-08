package scm

import (
	"context"
	"encoding/base64"

	"github.com/google/go-github/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type GithubAdaptor struct {
	Client *github.Client
}

func NewGithubAdaptor(ctx context.Context, pat string) ScmAdaptor {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	service := GithubAdaptor{
		Client: client,
	}

	return &service
}

func (c *GithubAdaptor) GetChangelog(ctx context.Context, owner string, repo string, fromTag string, toTag string) (*[]ScmCommit, error) {
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

func (c *GithubAdaptor) GetUserRepos(ctx context.Context, user string) ([]ScmRepository, error) {
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

	var allScmRepos []ScmRepository

	for _, repo := range allRepos {
		scmRepo := ScmRepository{
			DefaultBranch: *repo.DefaultBranch,
			Name:          *repo.Name,
			OwnerName:     *repo.Owner.Login,
		}
		allScmRepos = append(allScmRepos, scmRepo)
	}

	return allScmRepos, nil
}

func (c *GithubAdaptor) GetRepoBranch(ctx context.Context, owner string, repo string, branchName string) (*ScmBranch, error) {
	branches, _, err := c.Client.Repositories.ListBranches(ctx, owner, repo, nil)
	if err != nil {
		log.Error().Err(err).Msg("Could not get repo branches")
		return nil, err
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

func (c *GithubAdaptor) GetRepoFile(ctx context.Context, owner string, repo string, sha string, filePath string) ([]byte, error) {
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

			raw, err := base64.StdEncoding.DecodeString(*content.Content)
			if err != nil {
				log.Error().Err(err).Msg("Could not decode repo file")
				return nil, err
			}

			return raw, nil
		}
	}

	return nil, nil
}
