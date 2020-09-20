package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type GithubService interface {
	GetChangelog(ctx context.Context, owner string, repo string, fromTag string, toTag string) (*github.CommitsComparison, error)
	GetDashboardRepos(ctx context.Context) error
	GetRepoBranch(ctx context.Context, repo *github.Repository, branchName string) (*github.Branch, error)
	GetDashboardRepoConfig(ctx context.Context, owner string, repo string, sha string) (*dashboardRepoConfig, error)
	GetUserRepos(ctx context.Context, user string) ([]*github.Repository, error)
}

type githubService struct {
	Client *github.Client
}

type DashboardRepo struct {
	Repository *github.Repository
	Config     *dashboardRepoConfig
}

type dashboardRepoConfig struct {
	Name            string `yaml:"name"`
	GocdEnvironment string `yaml:"gocd_environment"`
	Pipeline        struct {
		Service []struct {
			CronEnvs  []string `yaml:"cron_envs"`
			CronTimer string   `yaml:"cron_timer"`
			Name      string   `yaml:"name"`
			RepoUrl   string   `yaml:"repo_url"`
			Whitelist []string `yaml:"whitelist"`
		} `yaml:"service"`
	} `yaml:"pipeline"`
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

func (c *githubService) GetDashboardRepos(ctx context.Context) error {
	allRepos, err := c.GetUserRepos(ctx, "")
	if err != nil {
		log.Println(err)
		return err
	}

	var dashboardRepos []DashboardRepo

	for _, repo := range allRepos {
		branch, err := c.GetRepoBranch(ctx, repo, "master")
		if err != nil {
			log.Println(err)
			continue
		}
		if branch == nil {
			continue
		}

		repoConfig, err := c.GetDashboardRepoConfig(ctx, *repo.Owner.Login, *repo.Name, *branch.Commit.SHA)
		if err != nil {
			log.Println(err)
			continue
		}
		if repoConfig == nil {
			continue
		}

		dashboardRepo := DashboardRepo{
			Config:     repoConfig,
			Repository: repo,
		}

		dashboardRepos = append(dashboardRepos, dashboardRepo)
	}

	fmt.Println(dashboardRepos)

	return nil
}

func (c *githubService) GetUserRepos(ctx context.Context, user string) ([]*github.Repository, error) {
	opts := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepos []*github.Repository
	for {
		repos, resp, err := c.Client.Repositories.List(ctx, "", opts)
		if err != nil {
			log.Println(err)
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

func (c *githubService) GetRepoBranch(ctx context.Context, repo *github.Repository, branchName string) (*github.Branch, error) {
	branches, _, err := c.Client.Repositories.ListBranches(ctx, *repo.Owner.Login, *repo.Name, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for _, branch := range branches {
		if *branch.Name == branchName {
			return branch, nil
		}
	}

	return nil, nil
}

func (c *githubService) GetDashboardRepoConfig(ctx context.Context, owner string, repo string, sha string) (*dashboardRepoConfig, error) {
	repoTree, _, err := c.Client.Git.GetTree(ctx, owner, repo, sha, true)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for _, treeEntry := range repoTree.Entries {
		if *treeEntry.Path == "deployment/smartshop.yml" {
			content, _, _, err := c.Client.Repositories.GetContents(ctx, owner, repo, "deployment/smartshop.yml", nil)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			repoConfig, err := NewDashboardRepoConfig(content)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			return repoConfig, nil
		}
	}

	return nil, nil
}

func NewDashboardRepoConfig(content *github.RepositoryContent) (*dashboardRepoConfig, error) {
	raw, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	repoConfig := dashboardRepoConfig{}

	err = yaml.Unmarshal([]byte(string(raw)), &repoConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}

	return &repoConfig, nil
}
