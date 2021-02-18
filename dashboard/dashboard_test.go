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

	mockRepos := []scm.ScmRepository{{
		DefaultBranch: "main",
		Name:          mockRepoName,
		OwnerName:     mockOwner,
	}}

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

func TestGetDashboardRepoConfigWithAllOptions(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockScm := mock_scm.NewMockScmAdapter(ctrl)
	dashboardService := dashboard.DashboardService{ScmService: mockScm}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepo := "r"
	mockDefaultBranch := "main"
	mockSha := "s"

	mockRepoBranch := scm.ScmRef{
		CurrentHash: mockSha,
		Name:        "main",
	}
	mockScm.
		EXPECT().
		GetRepoBranch(mockCtx, mockOwner, mockRepo, mockDefaultBranch).
		Times(1).
		Return(&mockRepoBranch, nil)

	mockConfigB64 := "LS0tCgplbnZpcm9ubWVudF9icmFuY2hlczoKICAtIHByZXByb2QKICAtIHByb2QKZW52aXJvbm1lbnRfdGFnczoKICAtIGRldgogIC0gcHJkCm5hbWU6IGFwcAo="
	mockRepoContent, _ := base64.StdEncoding.DecodeString(mockConfigB64)
	mockScm.
		EXPECT().
		GetRepoFile(mockCtx, mockOwner, mockRepo, mockSha, ".releasedash.yml").
		Times(1).
		Return(mockRepoContent, nil)

	mockConfig := dashboard.DashboardRepoConfig{
		EnvironmentBranches: []string{"preprod", "prod"},
		EnvironmentTags:     []string{"dev", "prd"},
		Name:                "app",
	}

	config, err := dashboardService.GetDashboardRepoConfig(mockCtx, mockOwner, mockRepo, mockDefaultBranch)
	assert.NoError(t, err)
	assert.Equal(t, &mockConfig, config)
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

	mockTagRepoName := "r-tag"
	mockTagRepo := scm.ScmRepository{
		DefaultBranch: "main",
		Name:          mockTagRepoName,
		OwnerName:     mockOwner,
	}
	mockTagCommitsCompare := []scm.ScmCommit{{Message: "m"}}

	mockBranchRepoName := "r-branch"
	mockBranchRepo := scm.ScmRepository{
		DefaultBranch: "main",
		Name:          mockBranchRepoName,
		OwnerName:     mockOwner,
	}
	mockBranchCommitsCompare := []scm.ScmCommit{{Message: "m"}}

	mockDashboardRepos := []dashboard.DashboardRepo{
		{
			Config: &dashboard.DashboardRepoConfig{
				EnvironmentBranches: []string{"pre-prod", "prod"},
				Name:                "app",
			},
			Repository: mockBranchRepo,
		},
		{
			Config: &dashboard.DashboardRepoConfig{
				EnvironmentTags: []string{"dev", "stg"},
				Name:            "app",
			},
			Repository: mockTagRepo,
		},
	}

	mockCtx := context.Background()
	mockScm.
		EXPECT().
		GetChangelogForBranches(mockCtx, mockOwner, mockBranchRepoName, "prod", "pre-prod").
		Times(1).
		Return(&mockBranchCommitsCompare, nil)
	mockScm.
		EXPECT().
		GetChangelogForTags(mockCtx, mockOwner, mockTagRepoName, "stg", "dev").
		Times(1).
		Return(&mockTagCommitsCompare, nil)

	repoChangelogs := dashboardService.GetDashboardChangelogs(mockCtx, mockDashboardRepos)

	expectedRepoChangelogs := []dashboard.DashboardRepoChangelog{
		{
			ChangelogCommits: []dashboard.DashboardChangelogCommits{{
				Commits: mockBranchCommitsCompare,
				FromRef: "prod",
				ToRef:   "pre-prod",
			}},
			Repository: mockBranchRepo,
		},
		{
			ChangelogCommits: []dashboard.DashboardChangelogCommits{{
				Commits: mockTagCommitsCompare,
				FromRef: "stg",
				ToRef:   "dev",
			}},
			Repository: mockTagRepo,
		},
	}

	assert.Equal(t, expectedRepoChangelogs, repoChangelogs)
}

func TestGetDashboardChangelogsNoChanges(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockScm := mock_scm.NewMockScmAdapter(ctrl)
	dashboardService := dashboard.DashboardService{ScmService: mockScm}

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepoName := "r"

	mockConfig := dashboard.DashboardRepoConfig{
		EnvironmentTags: []string{"dev", "stg"},
		Name:            "app",
	}

	mockDashboardRepos := []dashboard.DashboardRepo{{
		Config: &mockConfig,
		Repository: scm.ScmRepository{
			DefaultBranch: "main",
			Name:          mockRepoName,
			OwnerName:     mockOwner,
		},
	}}

	mockScm.
		EXPECT().
		GetChangelogForTags(mockCtx, mockOwner, mockRepoName, "stg", "dev").
		Times(1).
		Return(nil, errors.New(""))

	dashboardService.GetDashboardChangelogs(mockCtx, mockDashboardRepos)
}
