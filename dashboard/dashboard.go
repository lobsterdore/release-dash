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
	EnvironmentBranches []string `yaml:"environment_branches"`
	EnvironmentTags     []string `yaml:"environment_tags"`
	Name                string   `yaml:"name"`
}

type DashboardRepoChangelog struct {
	ChangelogCommits []DashboardChangelogCommits
	Repository       scm.ScmRepository
}

func (d DashboardRepoChangelog) HasChangelogCommits() bool {
	for _, changelogCommit := range d.ChangelogCommits {
		if len(changelogCommit.Commits) > 0 {
			return true
		}
	}
	return false
}

type DashboardChangelogCommits struct {
	Commits []scm.ScmCommit
	FromRef string
	ToRef   string
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

func (c *DashboardRepoConfig) HasEnvironmentBranches() bool {
	if c.EnvironmentBranches == nil || len(c.EnvironmentBranches) == 0 {
		return false
	}
	return true
}

func (c *DashboardRepoConfig) HasEnvironmentTags() bool {
	if c.EnvironmentTags == nil || len(c.EnvironmentTags) == 0 {
		return false
	}
	return true
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

		var environmentRefs []string
		repoConfig := dashboardRepo.Config
		if repoConfig.HasEnvironmentBranches() {
			environmentRefs = dashboardRepo.Config.EnvironmentBranches
		} else if repoConfig.HasEnvironmentTags() {
			environmentRefs = dashboardRepo.Config.EnvironmentTags
		} else {
			continue
		}

		for index, toRef := range environmentRefs {
			nextIndex := index + 1
			if nextIndex < len(environmentRefs) {
				fromRef := environmentRefs[nextIndex]
				log.Debug().Msgf("Getting changelog for tags %s - %s", fromRef, toRef)

				var changelog *[]scm.ScmCommit
				var err error

				if repoConfig.HasEnvironmentBranches() {
					changelog, err = d.ScmService.GetChangelogForBranches(ctx, org, repo, fromRef, toRef)
				} else {
					changelog, err = d.ScmService.GetChangelogForTags(ctx, org, repo, fromRef, toRef)
				}

				if err == nil {
					changelogCommits := DashboardChangelogCommits{
						FromRef: fromRef,
						ToRef:   toRef,
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
