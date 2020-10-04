package service_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	mock_service "github.com/lobsterdore/release-dash/mocks/service"
	service "github.com/lobsterdore/release-dash/service"
)

func TestGetDashboardChangelogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGithubService := mock_service.NewMockGithubService(ctrl)

	ctx := context.Background()
	owner := "ltfrankdrebin"
	repo := "policesquad"
	sha := "ringoffear"

	mockGithubService.
		EXPECT().
		GetRepoFile(ctx, owner, repo, sha, ".releasedash.yml").
		Times(1).
		Return(nil, nil)

	dashboard := service.DashboardService{GithubService: mockGithubService}

	_, err := dashboard.GetDashboardRepoConfig(ctx, owner, repo, sha)
	assert.NoError(t, err)
}
