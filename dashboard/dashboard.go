package dashboard

import (
	"context"
	"sort"
	"strings"

	"github.com/creasty/defaults"
	"github.com/lobsterdore/release-dash/config"
	"github.com/lobsterdore/release-dash/scm"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=-mod=mod --source=dashboard.go --destination=../mocks/dashboard/dashboard.go
type DashboardProvider interface {
	GetDashboardChangelogs(ctx context.Context, dashboardRepos []DashboardRepo) []DashboardRepoChangelog
	GetDashboardRepos(ctx context.Context) ([]DashboardRepo, error)
	GetDashboardRepoConfig(ctx context.Context, owner string, repo string, defaultBranch string) (*DashboardRepoConfig, error)
}

type DashboardService struct {
	ScmService scm.ScmAdapter
}

type DashboardRepo struct {
	Config     *DashboardRepoConfig
	Repository scm.ScmRepository
}

type DashboardRepoConfig struct {
	EnvironmentTags []string `yaml:"environment_tags"`
	Name            string   `yaml:"name"`
}

type DashboardRepoChangelog struct {
	ChangelogCommits []DashboardChangelogCommits
	Repository       scm.ScmRepository
}

type DashboardChangelogCommits struct {
	Commits []scm.ScmCommit
	FromTag string
	ToTag   string
}

func NewDashboardService(ctx context.Context, config config.Config, scmService scm.ScmAdapter) DashboardProvider {
	service := DashboardService{
		ScmService: scmService,
	}

	return &service
}

func NewDashboardRepoConfig(content []byte) (*DashboardRepoConfig, error) {
	repoConfig := &DashboardRepoConfig{}
	if err := defaults.Set(repoConfig); err != nil {
		log.Error().Err(err).Msg("Could not set repo config defaults")
		return nil, err
	}

	err := yaml.Unmarshal([]byte(string(content)), repoConfig)
	if err != nil {
		log.Error().Err(err).Msg("Could not set unmarshal repo config")
		return nil, err
	}

	return repoConfig, nil
}

func (d *DashboardService) GetDashboardRepos(ctx context.Context) ([]DashboardRepo, error) {
	allRepos, err := d.ScmService.GetUserRepos(ctx, "")
	if err != nil {
		log.Error().Err(err).Msg("Could not get dashboard repos")
		return nil, err
	}

	var dashboardRepos []DashboardRepo

	for _, repo := range allRepos {
		repoConfig, err := d.GetDashboardRepoConfig(ctx, repo.OwnerName, repo.Name, repo.DefaultBranch)
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

	sort.Slice(dashboardRepos, func(i, j int) bool {
		comparison := strings.Compare(dashboardRepos[i].Repository.Name, dashboardRepos[j].Repository.Name)
		return comparison != 1
	})

	return dashboardRepos, nil
}

func (d *DashboardService) GetDashboardRepoConfig(ctx context.Context, owner string, repo string, defaultBranch string) (*DashboardRepoConfig, error) {
	branch, err := d.ScmService.GetRepoBranch(ctx, owner, repo, defaultBranch)
	if err != nil {
		log.Error().Err(err).Msg("Could not get repo branch")
		return nil, nil
	}
	if branch == nil {
		return nil, nil
	}

	repoConfigContent, err := d.ScmService.GetRepoFile(ctx, owner, repo, branch.CurrentHash, ".releasedash.yml")
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
		org := dashboardRepo.Repository.OwnerName
		repo := dashboardRepo.Repository.Name

		repoChangelog := DashboardRepoChangelog{
			ChangelogCommits: []DashboardChangelogCommits{},
			Repository:       dashboardRepo.Repository,
		}

		environmentTags := dashboardRepo.Config.EnvironmentTags

		for index, toTag := range environmentTags {
			nextIndex := index + 1
			if nextIndex < len(environmentTags) {
				fromTag := environmentTags[nextIndex]
				changelog, err := d.ScmService.GetChangelog(ctx, org, repo, fromTag, toTag)
				if err == nil {
					changelogCommits := DashboardChangelogCommits{
						FromTag: fromTag,
						ToTag:   toTag,
					}
					if changelog != nil {
						changelogCommits.Commits = *changelog
					}
					repoChangelog.ChangelogCommits = append(repoChangelog.ChangelogCommits, changelogCommits)
				}
			}
		}
		repoChangelogs = append(repoChangelogs, repoChangelog)
	}

	return repoChangelogs
}
