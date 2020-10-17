package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"

	mock_service "github.com/lobsterdore/release-dash/mocks/service"
	"github.com/lobsterdore/release-dash/service"
	"github.com/lobsterdore/release-dash/web/handler"
)

func TestHomepageHasRepoHasChanges(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDashboardService := mock_service.NewMockDashboardProvider(ctrl)

	mockOwner := "o"
	mockRepoName := "r"
	mockAvatarURL := "au"

	mockUser := github.User{
		AvatarURL: &mockAvatarURL,
		Login:     &mockOwner,
	}
	mockRepo := github.Repository{
		Owner: &mockUser,
		Name:  &mockRepoName,
	}

	var mockDashboardRepos []service.DashboardRepo
	mockConfig := service.DashboardRepoConfig{
		EnvironmentTags: []string{"dev", "stg"},
		Name:            "app",
	}

	mockDashboardRepo := service.DashboardRepo{
		Config:     &mockConfig,
		Repository: &mockRepo,
	}

	mockDashboardRepos = append(mockDashboardRepos, mockDashboardRepo)

	mockCtx := context.Background()

	mockSha := "s"
	mockMessage := "mock message"
	mockCommit := github.Commit{
		SHA:     &mockSha,
		Message: &mockMessage,
	}

	mockUrl := "u"
	mockRepoCommit := github.RepositoryCommit{
		SHA:     &mockSha,
		Commit:  &mockCommit,
		HTMLURL: &mockUrl,
		Author:  &mockUser,
	}
	mockCommitsCompare := github.CommitsComparison{
		Commits: []github.RepositoryCommit{mockRepoCommit},
	}

	mockChangelogCommits := service.DashboardChangelogCommits{
		Commits: mockCommitsCompare.Commits,
		FromTag: "stg",
		ToTag:   "dev",
	}
	mockRepoChangelog := service.DashboardRepoChangelog{
		ChangelogCommits: []service.DashboardChangelogCommits{mockChangelogCommits},
		Repository:       mockRepo,
	}

	var mockRepoChangelogs []service.DashboardRepoChangelog
	mockRepoChangelogs = append(mockRepoChangelogs, mockRepoChangelog)

	mockDashboardService.
		EXPECT().
		GetDashboardChangelogs(mockCtx, mockDashboardRepos).
		Times(1).
		Return(mockRepoChangelogs)

	homepageHandler := handler.HomepageHandler{
		DashboardRepos:   mockDashboardRepos,
		DashboardService: mockDashboardService,
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req = req.WithContext(mockCtx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homepageHandler.Http)

	handler.ServeHTTP(rr, req)
	resBody := rr.Body.String()

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Contains(t, resBody, "<h2>"+mockRepoName+"</h2>")
	assert.Contains(t, resBody, "<h3>dev > stg</h3>")
	assert.Contains(t, resBody, mockMessage)
}

func TestHomepageHasRepoNoChanges(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockDashboardService := mock_service.NewMockDashboardProvider(ctrl)

	mockOwner := "o"
	mockRepoName := "r"

	mockUser := github.User{
		Login: &mockOwner,
	}
	mockRepo := github.Repository{
		Owner: &mockUser,
		Name:  &mockRepoName,
	}

	var mockDashboardRepos []service.DashboardRepo
	mockConfig := service.DashboardRepoConfig{
		EnvironmentTags: []string{"dev", "stg"},
		Name:            "app",
	}

	mockDashboardRepo := service.DashboardRepo{
		Config:     &mockConfig,
		Repository: &mockRepo,
	}

	mockDashboardRepos = append(mockDashboardRepos, mockDashboardRepo)

	mockCtx := context.Background()
	var mockRepoChangelogs []service.DashboardRepoChangelog

	mockDashboardService.
		EXPECT().
		GetDashboardChangelogs(mockCtx, mockDashboardRepos).
		Times(1).
		Return(mockRepoChangelogs)

	homepageHandler := handler.HomepageHandler{
		DashboardRepos:   mockDashboardRepos,
		DashboardService: mockDashboardService,
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req = req.WithContext(mockCtx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homepageHandler.Http)

	handler.ServeHTTP(rr, req)
	resBody := rr.Body.String()

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.NotContains(t, resBody, "<h2>r</h2>")
}
