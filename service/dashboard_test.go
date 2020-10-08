package service_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"

	mock_service "github.com/lobsterdore/release-dash/mocks/service"
	service "github.com/lobsterdore/release-dash/service"
)

func TestGetDashboardReposNoRepos(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithubService := mock_service.NewMockGithubService(ctrl)
	dashboard := service.DashboardService{GithubService: mockGithubService}

	mockCtx := context.Background()

	mockGithubService.
		EXPECT().
		GetUserRepos(mockCtx, "").
		Times(1).
		Return(nil, nil)

	repos, err := dashboard.GetDashboardRepos(mockCtx)

	var expectedRepos []service.DashboardRepo

	assert.NoError(t, err)
	assert.Equal(t, &expectedRepos, repos)
}

func TestGetDashboardReposNoConfigFiles(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithubService := mock_service.NewMockGithubService(ctrl)
	dashboard := service.DashboardService{GithubService: mockGithubService}

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

	repos, err := dashboard.GetDashboardRepos(mockCtx)

	var expectedRepos []service.DashboardRepo

	assert.NoError(t, err)
	assert.Equal(t, &expectedRepos, repos)
}

func TestGetDashboardReposHasRepos(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithubService := mock_service.NewMockGithubService(ctrl)
	dashboard := service.DashboardService{GithubService: mockGithubService}

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

	repos, err := dashboard.GetDashboardRepos(mockCtx)

	var expectedRepos []service.DashboardRepo
	mockConfig := service.DashboardRepoConfig{
		EnvironmentTags: []string{"dev"},
		Name:            "app",
	}

	expectedRepo := service.DashboardRepo{
		Config:     &mockConfig,
		Repository: &mockRepo,
	}

	expectedRepos = append(expectedRepos, expectedRepo)

	assert.NoError(t, err)
	assert.Equal(t, &expectedRepos, repos)
}

func TestGetDashboardRepoConfigNoBranch(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockGithubService := mock_service.NewMockGithubService(ctrl)
	dashboard := service.DashboardService{GithubService: mockGithubService}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepo := "r"

	mockGithubService.
		EXPECT().
		GetRepoBranch(mockCtx, mockOwner, mockRepo, "master").
		Times(1).
		Return(nil, nil)

	config, err := dashboard.GetDashboardRepoConfig(mockCtx, mockOwner, mockRepo)
	assert.NoError(t, err)
	assert.Nil(t, config)
}
