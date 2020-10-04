package service

import (
	"context"
	"encoding/base64"
	"log"

	"github.com/creasty/defaults"
	"github.com/google/go-github/github"
	"github.com/lobsterdore/release-dash/config"
	"gopkg.in/yaml.v2"
)

type DashboardService interface {
	GetDashboardChangelogs(ctx context.Context, dashboardRepos *[]DashboardRepo) []DashboardRepoChangelog
	GetDashboardRepos(ctx context.Context) (*[]DashboardRepo, error)
	GetDashboardRepoConfig(ctx context.Context, owner string, repo string, sha string) (*dashboardRepoConfig, error)
}

type dashboardService struct {
	GithubService GithubService
}

type DashboardRepo struct {
	Config     *dashboardRepoConfig
	Repository *github.Repository
}

type dashboardRepoConfig struct {
	EnvironmentTags []string `yaml:"environment_tags"`
	Name            string   `yaml:"name"`
}

type DashboardRepoChangelog struct {
	ChangelogCommits []dashboardChangelogCommits
	Repository       github.Repository
}

type dashboardChangelogCommits struct {
	Commits []github.RepositoryCommit
	FromTag string
	ToTag   string
}

func NewDashboardService(ctx context.Context, config config.Config) DashboardService {
	githubService := NewGithubService(ctx, config.Github.Pat)

	service := dashboardService{
		GithubService: githubService,
	}

	return &service
}

func NewDashboardRepoConfig(content *github.RepositoryContent) (*dashboardRepoConfig, error) {
	raw, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	repoConfig := &dashboardRepoConfig{}
	if err := defaults.Set(repoConfig); err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}

	err = yaml.Unmarshal([]byte(string(raw)), repoConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}

	return repoConfig, nil
}

func (d *dashboardService) GetDashboardRepos(ctx context.Context) (*[]DashboardRepo, error) {
	allRepos, err := d.GithubService.GetUserRepos(ctx, "")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var dashboardRepos []DashboardRepo

	for _, repo := range allRepos {
		branch, err := d.GithubService.GetRepoBranch(ctx, repo, "master")
		if err != nil {
			log.Println(err)
			continue
		}
		if branch == nil {
			continue
		}

		repoConfig, err := d.GetDashboardRepoConfig(ctx, *repo.Owner.Login, *repo.Name, *branch.Commit.SHA)
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

	return &dashboardRepos, nil
}

func (d *dashboardService) GetDashboardRepoConfig(ctx context.Context, owner string, repo string, sha string) (*dashboardRepoConfig, error) {
	repoConfigContent, err := d.GithubService.GetRepoFile(ctx, owner, repo, sha, ".releasedash.yml")
	if err != nil {
		return nil, err
	}
	if repoConfigContent == nil {
		return nil, nil
	}

	repoConfig, err := NewDashboardRepoConfig(repoConfigContent)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return repoConfig, nil
}

func (d *dashboardService) GetDashboardChangelogs(ctx context.Context, dashboardRepos *[]DashboardRepo) []DashboardRepoChangelog {

	var repoChangelogs []DashboardRepoChangelog
	for _, dashboardRepo := range *dashboardRepos {
		org := *dashboardRepo.Repository.Owner.Login
		repo := *dashboardRepo.Repository.Name

		repoChangelog := DashboardRepoChangelog{
			ChangelogCommits: []dashboardChangelogCommits{},
			Repository:       *dashboardRepo.Repository,
		}

		environmentTags := dashboardRepo.Config.EnvironmentTags

		for index, toTag := range environmentTags {
			nextIndex := index + 1
			if nextIndex < len(environmentTags) {
				fromTag := environmentTags[nextIndex]
				changelog, err := d.GithubService.GetChangelog(ctx, org, repo, fromTag, toTag)
				if err == nil {
					changelogCommits := dashboardChangelogCommits{
						Commits: changelog.Commits,
						FromTag: fromTag,
						ToTag:   toTag,
					}
					repoChangelog.ChangelogCommits = append(repoChangelog.ChangelogCommits, changelogCommits)

				}
			}
		}
		repoChangelogs = append(repoChangelogs, repoChangelog)
	}

	return repoChangelogs
}
