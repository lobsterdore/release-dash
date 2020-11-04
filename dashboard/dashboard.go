package dashboard

import (
	"context"
	"encoding/base64"

	"github.com/creasty/defaults"
	"github.com/google/go-github/github"
	"github.com/lobsterdore/release-dash/config"
	"github.com/lobsterdore/release-dash/scm"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=-mod=mod --source=dashboard.go --destination=../mocks/dashboard/dashboard.go
type DashboardProvider interface {
	GetDashboardChangelogs(ctx context.Context, dashboardRepos []DashboardRepo) []DashboardRepoChangelog
	GetDashboardRepos(ctx context.Context) ([]DashboardRepo, error)
	GetDashboardRepoConfig(ctx context.Context, owner string, repo string) (*DashboardRepoConfig, error)
}

type DashboardService struct {
	GithubService scm.GithubProvider
}

type DashboardRepo struct {
	Config     *DashboardRepoConfig
	Repository *github.Repository
}

type DashboardRepoConfig struct {
	EnvironmentTags []string `yaml:"environment_tags"`
	Name            string   `yaml:"name"`
}

type DashboardRepoChangelog struct {
	ChangelogCommits []DashboardChangelogCommits
	Repository       github.Repository
}

type DashboardChangelogCommits struct {
	Commits []github.RepositoryCommit
	FromTag string
	ToTag   string
}

func NewDashboardService(ctx context.Context, config config.Config) DashboardProvider {
	githubService := scm.NewGithubService(ctx, config.Github.Pat)

	service := DashboardService{
		GithubService: githubService,
	}

	return &service
}

func NewDashboardRepoConfig(content *github.RepositoryContent) (*DashboardRepoConfig, error) {
	raw, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		log.Error().Err(err).Msg("Could not decode repo config")
		return nil, err
	}

	repoConfig := &DashboardRepoConfig{}
	if err := defaults.Set(repoConfig); err != nil {
		log.Error().Err(err).Msg("Could not set repo config defaults")
		return nil, err
	}

	err = yaml.Unmarshal([]byte(string(raw)), repoConfig)
	if err != nil {
		log.Error().Err(err).Msg("Could not set unmarshal repo config")
		return nil, err
	}

	return repoConfig, nil
}

func (d *DashboardService) GetDashboardRepos(ctx context.Context) ([]DashboardRepo, error) {
	allRepos, err := d.GithubService.GetUserRepos(ctx, "")
	if err != nil {
		log.Error().Err(err).Msg("Could not get dashboard repos")
		return nil, err
	}

	var dashboardRepos []DashboardRepo

	for _, repo := range allRepos {
		repoConfig, err := d.GetDashboardRepoConfig(ctx, *repo.Owner.Login, *repo.Name)
		if err != nil {
			log.Error().Err(err).Msg("Could not get repo config file")
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

	return dashboardRepos, nil
}

func (d *DashboardService) GetDashboardRepoConfig(ctx context.Context, owner string, repo string) (*DashboardRepoConfig, error) {
	branch, err := d.GithubService.GetRepoBranch(ctx, owner, repo, "master")
	if err != nil {
		log.Error().Err(err).Msg("Could not get repo branch")
		return nil, nil
	}
	if branch == nil {
		return nil, nil
	}

	repoConfigContent, err := d.GithubService.GetRepoFile(ctx, owner, repo, *branch.Commit.SHA, ".releasedash.yml")
	if err != nil {
		return nil, err
	}
	if repoConfigContent == nil {
		return nil, nil
	}

	repoConfig, err := NewDashboardRepoConfig(repoConfigContent)
	if err != nil {
		log.Error().Err(err).Msg("Could not create repo config")
		return nil, err
	}
	return repoConfig, nil
}

func (d *DashboardService) GetDashboardChangelogs(ctx context.Context, dashboardRepos []DashboardRepo) []DashboardRepoChangelog {

	var repoChangelogs []DashboardRepoChangelog
	for _, dashboardRepo := range dashboardRepos {
		org := *dashboardRepo.Repository.Owner.Login
		repo := *dashboardRepo.Repository.Name

		repoChangelog := DashboardRepoChangelog{
			ChangelogCommits: []DashboardChangelogCommits{},
			Repository:       *dashboardRepo.Repository,
		}

		environmentTags := dashboardRepo.Config.EnvironmentTags

		for index, toTag := range environmentTags {
			nextIndex := index + 1
			if nextIndex < len(environmentTags) {
				fromTag := environmentTags[nextIndex]
				changelog, err := d.GithubService.GetChangelog(ctx, org, repo, fromTag, toTag)
				if err == nil {
					changelogCommits := DashboardChangelogCommits{
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