package dashboard

import (
	"context"
	"fmt"
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

func NewDashboardService(ctx context.Context, config config.Config, scmService scm.ScmAdapter) *DashboardService {
	service := DashboardService{
		ScmService: scmService,
	}
	return &service
}

func NewDashboardRepoConfig(content []byte) (*DashboardRepoConfig, error) {
	repoConfig := &DashboardRepoConfig{}
	if err := defaults.Set(repoConfig); err != nil {
		return nil, fmt.Errorf("Could not set repo config defaults: %s", err)
	}

	err := yaml.Unmarshal([]byte(string(content)), repoConfig)
	if err != nil {
		return nil, fmt.Errorf("Could not set unmarshal repo config: %s", err)
	}

	return repoConfig, nil
}

func (d *DashboardService) GetDashboardRepos(ctx context.Context) ([]DashboardRepo, error) {
	allRepos, err := d.ScmService.GetUserRepos(ctx, "")
	if err != nil {
		return nil, err
	}

	var dashboardRepos []DashboardRepo

	for _, repo := range allRepos {
		log.Debug().Msgf("Checking repo %s/%s for config file", repo.OwnerName, repo.Name)
		repoConfig, err := d.GetDashboardRepoConfig(ctx, repo.OwnerName, repo.Name, repo.DefaultBranch)
		if err != nil {
			log.Error().Err(err).Msgf("Could not get repo config file %s/%s", repo.OwnerName, repo.Name)
			continue
		}
		if repoConfig == nil {
			log.Debug().Msgf("No config file for repo %s/%s", repo.OwnerName, repo.Name)
			continue
		}

		dashboardRepo := DashboardRepo{
			Config:     repoConfig,
			Repository: repo,
		}

		dashboardRepos = append(dashboardRepos, dashboardRepo)
		log.Debug().Msgf("Repo %s/%s added to dashboard", repo.OwnerName, repo.Name)
	}

	sort.Slice(dashboardRepos, func(i, j int) bool {
		comparison := strings.Compare(dashboardRepos[i].Repository.Name, dashboardRepos[j].Repository.Name)
		return comparison != 1
	})

	return dashboardRepos, nil
}

func (d *DashboardService) GetDashboardRepoConfig(ctx context.Context, owner string, repo string, defaultBranch string) (*DashboardRepoConfig, error) {
	configFilePath := ".releasedash.yml"
	branch, err := d.ScmService.GetRepoBranch(ctx, owner, repo, defaultBranch)
	if err != nil {
		log.Error().Err(err).Msgf("Could not get repo %s/%s branch %s", owner, repo, defaultBranch)
		return nil, nil
	}
	if branch == nil {
		log.Debug().Msgf("Repo %s/%s does not have branch %s", owner, repo, defaultBranch)
		return nil, nil
	}

	repoConfigContent, err := d.ScmService.GetRepoFile(ctx, owner, repo, branch.CurrentHash, configFilePath)
	if err != nil {
		return nil, err
	}
	if repoConfigContent == nil {
		log.Debug().Msgf("Repo %s/%s does not have file %s", owner, repo, configFilePath)
		return nil, nil
	}

	repoConfig, err := NewDashboardRepoConfig(repoConfigContent)
	if err != nil {
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

		log.Debug().Msgf("Getting changelog for Repo %s/%s", org, repo)

		environmentTags := dashboardRepo.Config.EnvironmentTags

		for index, toTag := range environmentTags {
			nextIndex := index + 1
			if nextIndex < len(environmentTags) {
				fromTag := environmentTags[nextIndex]
				log.Debug().Msgf("Getting changelog for tags %s - %s", fromTag, toTag)

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
				} else {
					log.Error().Err(err).Msg("Could not get changelog")
				}
			}
		}
		repoChangelogs = append(repoChangelogs, repoChangelog)
	}

	return repoChangelogs
}
