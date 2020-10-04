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

func TestGetDashboardRepoConfigHasFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGithubService := mock_service.NewMockGithubService(ctrl)

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepo := "r"
	mockSha := "s"

	mockConfigB64 := "LS0tCgplbnZpcm9ubWVudF90YWdzOgogIC0gZGV2Cm5hbWU6IGFwcAo="
	mockRepoContent := github.RepositoryContent{
		Content: &mockConfigB64,
	}

	expectedConfig := service.DashboardRepoConfig{
		EnvironmentTags: []string{"dev"},
		Name:            "app",
	}

	mockGithubService.
		EXPECT().
		GetRepoFile(mockCtx, mockOwner, mockRepo, mockSha, ".releasedash.yml").
		Times(1).
		Return(&mockRepoContent, nil)

	dashboard := service.DashboardService{GithubService: mockGithubService}

	config, err := dashboard.GetDashboardRepoConfig(mockCtx, mockOwner, mockRepo, mockSha)
	assert.NoError(t, err)

	assert.Equal(t, &expectedConfig, config)
}

func TestGetDashboardRepoConfigNoFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGithubService := mock_service.NewMockGithubService(ctrl)

	mockCtx := context.Background()
	mockOwner := "o"
	mockRepo := "r"
	mockSha := "s"

	mockGithubService.
		EXPECT().
		GetRepoFile(mockCtx, mockOwner, mockRepo, mockSha, ".releasedash.yml").
		Times(1).
		Return(nil, nil)

	dashboard := service.DashboardService{GithubService: mockGithubService}

	config, err := dashboard.GetDashboardRepoConfig(mockCtx, mockOwner, mockRepo, mockSha)
	assert.NoError(t, err)
	assert.Nil(t, config)
}
