package dashboard_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	dashboard "github.com/lobsterdore/release-dash/dashboard"
	mock_scm "github.com/lobsterdore/release-dash/mocks/scm"
	"github.com/lobsterdore/release-dash/scm"
)

func TestGetDashboardReposNoRepos(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockScm := mock_scm.NewMockScmAdapter(ctrl)
	dashboardService := dashboard.DashboardService{ScmService: mockScm}

	mockCtx := context.Background()

	mockScm.
		EXPECT().
		GetUserRepos(mockCtx, "").
		Times(1).
		Return(nil, nil)

	repos, err := dashboardService.GetDashboardRepos(mockCtx)

	var expectedRepos []dashboard.DashboardRepo

	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}

func TestGetDashboardReposError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockScm := mock_scm.NewMockScmAdapter(ctrl)
	dashboardService := dashboard.DashboardService{ScmService: mockScm}

	mockCtx := context.Background()

	mockScm.
		EXPECT().
		GetUserRepos(mockCtx, "").
		Times(1).
		Return(nil, errors.New("Error"))

	repos, err := dashboardService.GetDashboardRepos(mockCtx)

	assert.Error(t, err)
	assert.Nil(t, repos)
}

func TestGetDashboardReposNoConfigFiles(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockScm := mock_scm.NewMockScmAdapter(ctrl)
	dashboardService := dashboard.DashboardService{ScmService: mockScm}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepoName := "r"
	mockSha := "s"

	var mockRepos []scm.ScmRepository

	mockRepo := scm.ScmRepository{
		DefaultBranch: "main",
		Name:          mockRepoName,
		OwnerName:     mockOwner,
	}
	mockRepos = append(mockRepos, mockRepo)

	mockScm.
		EXPECT().
		GetUserRepos(mockCtx, "").
		Times(1).
		Return(mockRepos, nil)

	mockRepoBranch := scm.ScmRef{
		CurrentHash: mockSha,
		Name:        "main",
	}
	mockScm.
		EXPECT().
		GetRepoBranch(mockCtx, mockOwner, mockRepoName, "main").
		Times(1).
		Return(&mockRepoBranch, nil)

	mockScm.
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

	mockScm := mock_scm.NewMockScmAdapter(ctrl)
	dashboardService := dashboard.DashboardService{ScmService: mockScm}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepoName := "r"
	mockSha := "s"

	mockRepoA := scm.ScmRepository{
		DefaultBranch: "main",
		Name:          mockRepoName + "a",
		OwnerName:     mockOwner,
	}
	mockRepoB := scm.ScmRepository{
		DefaultBranch: "main",
		Name:          mockRepoName + "b",
		OwnerName:     mockOwner,
	}
	mockRepos := []scm.ScmRepository{mockRepoB, mockRepoA}

	mockScm.
		EXPECT().
		GetUserRepos(mockCtx, "").
		Times(1).
		Return(mockRepos, nil)

	mockRepoBranch := scm.ScmRef{
		CurrentHash: mockSha,
		Name:        "main",
	}
	mockScm.
		EXPECT().
		GetRepoBranch(mockCtx, mockOwner, mockRepoName+"a", "main").
		Times(1).
		Return(&mockRepoBranch, nil)
	mockScm.
		EXPECT().
		GetRepoBranch(mockCtx, mockOwner, mockRepoName+"b", "main").
		Times(1).
		Return(&mockRepoBranch, nil)

	mockConfigB64 := "LS0tCgplbnZpcm9ubWVudF90YWdzOgogIC0gZGV2Cm5hbWU6IGFwcAo="
	mockRepoContent, _ := base64.StdEncoding.DecodeString(mockConfigB64)
	mockScm.
		EXPECT().
		GetRepoFile(mockCtx, mockOwner, mockRepoName+"a", mockSha, ".releasedash.yml").
		Times(1).
		Return(mockRepoContent, nil)
	mockScm.
		EXPECT().
		GetRepoFile(mockCtx, mockOwner, mockRepoName+"b", mockSha, ".releasedash.yml").
		Times(1).
		Return(mockRepoContent, nil)

	repos, err := dashboardService.GetDashboardRepos(mockCtx)

	mockConfig := dashboard.DashboardRepoConfig{
		EnvironmentTags: []string{"dev"},
		Name:            "app",
	}

	expectedRepos := []dashboard.DashboardRepo{
		{
			Config:     &mockConfig,
			Repository: mockRepoA,
		},
		{
			Config:     &mockConfig,
			Repository: mockRepoB,
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}

func TestGetDashboardReposBadConfigFile(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockScm := mock_scm.NewMockScmAdapter(ctrl)
	dashboardService := dashboard.DashboardService{ScmService: mockScm}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepoName := "r"
	mockSha := "s"

	mockRepoA := scm.ScmRepository{
		DefaultBranch: "main",
		Name:          mockRepoName + "a",
		OwnerName:     mockOwner,
	}
	mockRepoB := scm.ScmRepository{
		DefaultBranch: "main",
		Name:          mockRepoName + "b",
		OwnerName:     mockOwner,
	}
	mockRepos := []scm.ScmRepository{mockRepoB, mockRepoA}

	mockScm.
		EXPECT().
		GetUserRepos(mockCtx, "").
		Times(1).
		Return(mockRepos, nil)

	mockRepoBranch := scm.ScmRef{
		CurrentHash: mockSha,
		Name:        "main",
	}
	mockScm.
		EXPECT().
		GetRepoBranch(mockCtx, mockOwner, mockRepoName+"a", "main").
		Times(1).
		Return(&mockRepoBranch, nil)
	mockScm.
		EXPECT().
		GetRepoBranch(mockCtx, mockOwner, mockRepoName+"b", "main").
		Times(1).
		Return(&mockRepoBranch, nil)

	mockGoodConfigB64 := "LS0tCgplbnZpcm9ubWVudF90YWdzOgogIC0gZGV2Cm5hbWU6IGFwcAo="
	mockGoodRepoContent, _ := base64.StdEncoding.DecodeString(mockGoodConfigB64)

	mockBadConfigB64 := "LS0tCgplbnZpcm9ubWVudAo="
	mockBadRepoContent, _ := base64.StdEncoding.DecodeString(mockBadConfigB64)

	mockScm.
		EXPECT().
		GetRepoFile(mockCtx, mockOwner, mockRepoName+"a", mockSha, ".releasedash.yml").
		Times(1).
		Return(mockGoodRepoContent, nil)
	mockScm.
		EXPECT().
		GetRepoFile(mockCtx, mockOwner, mockRepoName+"b", mockSha, ".releasedash.yml").
		Times(1).
		Return(mockBadRepoContent, nil)

	repos, err := dashboardService.GetDashboardRepos(mockCtx)

	mockConfig := dashboard.DashboardRepoConfig{
		EnvironmentTags: []string{"dev"},
		Name:            "app",
	}

	expectedRepos := []dashboard.DashboardRepo{{
		Config:     &mockConfig,
		Repository: mockRepoA,
	}}

	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}

func TestGetDashboardRepoConfigNoBranch(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockScm := mock_scm.NewMockScmAdapter(ctrl)
	dashboardService := dashboard.DashboardService{ScmService: mockScm}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepo := "r"
	mockDefaultBranch := "main"

	mockScm.
		EXPECT().
		GetRepoBranch(mockCtx, mockOwner, mockRepo, mockDefaultBranch).
		Times(1).
		Return(nil, nil)

	config, err := dashboardService.GetDashboardRepoConfig(mockCtx, mockOwner, mockRepo, mockDefaultBranch)
	assert.NoError(t, err)
	assert.Nil(t, config)
}

func TestGetDashboardChangelogsHasChanges(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockScm := mock_scm.NewMockScmAdapter(ctrl)
	dashboardService := dashboard.DashboardService{ScmService: mockScm}

	mockOwner := "o"
	mockRepoName := "r"

	mockRepo := scm.ScmRepository{
		DefaultBranch: "main",
		Name:          mockRepoName,
		OwnerName:     mockOwner,
	}

	var mockDashboardRepos []dashboard.DashboardRepo
	mockConfig := dashboard.DashboardRepoConfig{
		EnvironmentTags: []string{"dev", "stg"},
		Name:            "app",
	}

	mockDashboardRepo := dashboard.DashboardRepo{
		Config:     &mockConfig,
		Repository: mockRepo,
	}

	mockDashboardRepos = append(mockDashboardRepos, mockDashboardRepo)

	mockCommit := scm.ScmCommit{Message: "m"}
	mockCommitsCompare := []scm.ScmCommit{mockCommit}

	mockCtx := context.Background()
	mockScm.
		EXPECT().
		GetChangelog(mockCtx, mockOwner, mockRepoName, "stg", "dev").
		Times(1).
		Return(&mockCommitsCompare, nil)

	expectedChangelogCommits := dashboard.DashboardChangelogCommits{
		Commits: mockCommitsCompare,
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

	mockScm := mock_scm.NewMockScmAdapter(ctrl)
	dashboardService := dashboard.DashboardService{ScmService: mockScm}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepoName := "r"

	mockRepo := scm.ScmRepository{
		DefaultBranch: "main",
		Name:          mockRepoName,
		OwnerName:     mockOwner,
	}

	var mockDashboardRepos []dashboard.DashboardRepo
	mockConfig := dashboard.DashboardRepoConfig{
		EnvironmentTags: []string{"dev", "stg"},
		Name:            "app",
	}

	mockDashboardRepo := dashboard.DashboardRepo{
		Config:     &mockConfig,
		Repository: mockRepo,
	}

	mockDashboardRepos = append(mockDashboardRepos, mockDashboardRepo)

	mockScm.
		EXPECT().
		GetChangelog(mockCtx, mockOwner, mockRepoName, "stg", "dev").
		Times(1).
		Return(nil, errors.New(""))

	dashboardService.GetDashboardChangelogs(mockCtx, mockDashboardRepos)
}
