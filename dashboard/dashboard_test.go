package dashboard_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"

	dashboard "github.com/lobsterdore/release-dash/dashboard"
	mock_scm "github.com/lobsterdore/release-dash/mocks/scm"
)

func TestGetDashboardReposNoRepos(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithubService := mock_scm.NewMockGithubProvider(ctrl)
	dashboardService := dashboard.DashboardService{GithubService: mockGithubService}

	mockCtx := context.Background()

	mockGithubService.
		EXPECT().
		GetUserRepos(mockCtx, "").
		Times(1).
		Return(nil, nil)

	repos, err := dashboardService.GetDashboardRepos(mockCtx)

	var expectedRepos []dashboard.DashboardRepo

	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}

func TestGetDashboardReposNoConfigFiles(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithubService := mock_scm.NewMockGithubProvider(ctrl)
	dashboardService := dashboard.DashboardService{GithubService: mockGithubService}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepoName := "r"
	mockSha := "s"

	var mockRepos []*github.Repository

	mockUser := github.User{
		Login: &mockOwner,
	}
	mockRepo := github.Repository{
		Owner: &mockUser,
		Name:  &mockRepoName,
	}
	mockRepos = append(mockRepos, &mockRepo)

	mockGithubService.
		EXPECT().
		GetUserRepos(mockCtx, "").
		Times(1).
		Return(mockRepos, nil)

	mockCommit := github.RepositoryCommit{
		SHA: &mockSha,
	}
	mockRepoBranch := github.Branch{
		Commit: &mockCommit,
	}
	mockGithubService.
		EXPECT().
		GetRepoBranch(mockCtx, mockOwner, mockRepoName, "master").
		Times(1).
		Return(&mockRepoBranch, nil)

	mockGithubService.
		EXPECT().
		GetRepoFile(mockCtx, mockOwner, mockRepoName, mockSha, ".releasedash.yml").
		Times(1).
		Return(nil, nil)

	repos, err := dashboardService.GetDashboardRepos(mockCtx)

	var expectedRepos []dashboard.DashboardRepo

	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}

func TestGetDashboardReposHasRepos(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithubService := mock_scm.NewMockGithubProvider(ctrl)
	dashboardService := dashboard.DashboardService{GithubService: mockGithubService}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepoName := "r"
	mockSha := "s"

	var mockRepos []*github.Repository

	mockUser := github.User{
		Login: &mockOwner,
	}
	mockRepo := github.Repository{
		Owner: &mockUser,
		Name:  &mockRepoName,
	}
	mockRepos = append(mockRepos, &mockRepo)

	mockGithubService.
		EXPECT().
		GetUserRepos(mockCtx, "").
		Times(1).
		Return(mockRepos, nil)

	mockCommit := github.RepositoryCommit{
		SHA: &mockSha,
	}
	mockRepoBranch := github.Branch{
		Commit: &mockCommit,
	}
	mockGithubService.
		EXPECT().
		GetRepoBranch(mockCtx, mockOwner, mockRepoName, "master").
		Times(1).
		Return(&mockRepoBranch, nil)

	mockConfigB64 := "LS0tCgplbnZpcm9ubWVudF90YWdzOgogIC0gZGV2Cm5hbWU6IGFwcAo="
	mockRepoContent := github.RepositoryContent{
		Content: &mockConfigB64,
	}
	mockGithubService.
		EXPECT().
		GetRepoFile(mockCtx, mockOwner, mockRepoName, mockSha, ".releasedash.yml").
		Times(1).
		Return(&mockRepoContent, nil)

	repos, err := dashboardService.GetDashboardRepos(mockCtx)

	var expectedRepos []dashboard.DashboardRepo
	mockConfig := dashboard.DashboardRepoConfig{
		EnvironmentTags: []string{"dev"},
		Name:            "app",
	}

	expectedRepo := dashboard.DashboardRepo{
		Config:     &mockConfig,
		Repository: &mockRepo,
	}

	expectedRepos = append(expectedRepos, expectedRepo)

	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}

func TestGetDashboardRepoConfigNoBranch(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithubService := mock_scm.NewMockGithubProvider(ctrl)
	dashboardService := dashboard.DashboardService{GithubService: mockGithubService}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepo := "r"

	mockGithubService.
		EXPECT().
		GetRepoBranch(mockCtx, mockOwner, mockRepo, "master").
		Times(1).
		Return(nil, nil)

	config, err := dashboardService.GetDashboardRepoConfig(mockCtx, mockOwner, mockRepo)
	assert.NoError(t, err)
	assert.Nil(t, config)
}

func TestGetDashboardChangelogsHasChanges(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithubService := mock_scm.NewMockGithubProvider(ctrl)
	dashboardService := dashboard.DashboardService{GithubService: mockGithubService}

	mockOwner := "o"
	mockRepoName := "r"

	mockUser := github.User{
		Login: &mockOwner,
	}
	mockRepo := github.Repository{
		Owner: &mockUser,
		Name:  &mockRepoName,
	}

	var mockDashboardRepos []dashboard.DashboardRepo
	mockConfig := dashboard.DashboardRepoConfig{
		EnvironmentTags: []string{"dev", "stg"},
		Name:            "app",
	}

	mockDashboardRepo := dashboard.DashboardRepo{
		Config:     &mockConfig,
		Repository: &mockRepo,
	}

	mockDashboardRepos = append(mockDashboardRepos, mockDashboardRepo)

	mockSha := "s"
	mockCommit := github.RepositoryCommit{SHA: &mockSha}
	mockCommitsCompare := github.CommitsComparison{
		Commits: []github.RepositoryCommit{mockCommit},
	}

	mockCtx := context.Background()
	mockGithubService.
		EXPECT().
		GetChangelog(mockCtx, mockOwner, mockRepoName, "stg", "dev").
		Times(1).
		Return(&mockCommitsCompare, nil)

	expectedChangelogCommits := dashboard.DashboardChangelogCommits{
		Commits: mockCommitsCompare.Commits,
		FromTag: "stg",
		ToTag:   "dev",
	}
	expectedRepoChangelog := dashboard.DashboardRepoChangelog{
		ChangelogCommits: []dashboard.DashboardChangelogCommits{expectedChangelogCommits},
		Repository:       mockRepo,
	}

	expectedRepoChangelogs := []dashboard.DashboardRepoChangelog{expectedRepoChangelog}

	repoChangelogs := dashboardService.GetDashboardChangelogs(mockCtx, mockDashboardRepos)

	assert.Equal(t, expectedRepoChangelogs, repoChangelogs)
}

func TestGetDashboardChangelogsNoChanges(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithubService := mock_scm.NewMockGithubProvider(ctrl)
	dashboardService := dashboard.DashboardService{GithubService: mockGithubService}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepoName := "r"

	mockUser := github.User{
		Login: &mockOwner,
	}
	mockRepo := github.Repository{
		Owner: &mockUser,
		Name:  &mockRepoName,
	}

	var mockDashboardRepos []dashboard.DashboardRepo
	mockConfig := dashboard.DashboardRepoConfig{
		EnvironmentTags: []string{"dev", "stg"},
		Name:            "app",
	}

	mockDashboardRepo := dashboard.DashboardRepo{
		Config:     &mockConfig,
		Repository: &mockRepo,
	}

	mockDashboardRepos = append(mockDashboardRepos, mockDashboardRepo)

	mockGithubService.
		EXPECT().
		GetChangelog(mockCtx, mockOwner, mockRepoName, "stg", "dev").
		Times(1).
		Return(nil, errors.New(""))

	dashboardService.GetDashboardChangelogs(mockCtx, mockDashboardRepos)
}
