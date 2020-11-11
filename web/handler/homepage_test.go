package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"

	"github.com/lobsterdore/release-dash/dashboard"
	"github.com/lobsterdore/release-dash/scm"
	"github.com/lobsterdore/release-dash/web/handler"

	mock_cache "github.com/lobsterdore/release-dash/mocks/cache"
	mock_dashboard "github.com/lobsterdore/release-dash/mocks/dashboard"
)

func TestHomepageHasRepoHasChanges(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCacheService := mock_cache.NewMockCacheAdapter(ctrl)
	mockDashboardService := mock_dashboard.NewMockDashboardProvider(ctrl)

	mockOwner := "o"
	mockRepoName := "r"
	mockAvatarURL := "au"

	mockRepo := scm.ScmRepository{
		OwnerName: mockOwner,
		Name:      mockRepoName,
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

	mockCtx := context.Background()

	mockMessage := "mock message"
	mockUrl := "u"
	mockCommit := scm.ScmCommit{
		AuthorAvatarUrl: mockAvatarURL,
		Message:         mockMessage,
		HtmlUrl:         mockUrl,
	}
	mockCommitsCompare := []scm.ScmCommit{mockCommit}

	mockChangelogCommits := dashboard.DashboardChangelogCommits{
		Commits: mockCommitsCompare,
		FromTag: "stg",
		ToTag:   "dev",
	}
	mockRepoChangelog := dashboard.DashboardRepoChangelog{
		ChangelogCommits: []dashboard.DashboardChangelogCommits{mockChangelogCommits},
		Repository:       mockRepo,
	}

	var mockRepoChangelogs []dashboard.DashboardRepoChangelog
	mockRepoChangelogs = append(mockRepoChangelogs, mockRepoChangelog)

	mockHomepageData := handler.HomepageData{
		RepoChangelogs: mockRepoChangelogs,
	}

	mockCacheService.
		EXPECT().
		Get("homepage_data").
		Times(1).
		Return(nil, false)

	mockDashboardService.
		EXPECT().
		GetDashboardChangelogs(mockCtx, mockDashboardRepos).
		Times(1).
		Return(mockRepoChangelogs)

	mockCacheService.
		EXPECT().
		Set("homepage_data", mockHomepageData).
		Times(1)

	homepageHandler := handler.HomepageHandler{
		CacheService:     mockCacheService,
		DashboardRepos:   mockDashboardRepos,
		DashboardService: mockDashboardService,
		HasDashboardData: true,
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
	assert.Contains(t, resBody, "dev > stg")
	assert.Contains(t, resBody, mockMessage)
}

func TestHomepageHasRepoNoChanges(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCacheService := mock_cache.NewMockCacheAdapter(ctrl)
	mockDashboardService := mock_dashboard.NewMockDashboardProvider(ctrl)

	mockOwner := "o"
	mockRepoName := "r"

	mockRepo := scm.ScmRepository{
		Name:      mockRepoName,
		OwnerName: mockOwner,
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

	mockCtx := context.Background()
	var mockRepoChangelogs []dashboard.DashboardRepoChangelog

	mockHomepageData := handler.HomepageData{
		RepoChangelogs: mockRepoChangelogs,
	}

	mockCacheService.
		EXPECT().
		Get("homepage_data").
		Times(1).
		Return(nil, false)

	mockDashboardService.
		EXPECT().
		GetDashboardChangelogs(mockCtx, mockDashboardRepos).
		Times(1).
		Return(mockRepoChangelogs)

	mockCacheService.
		EXPECT().
		Set("homepage_data", mockHomepageData).
		Times(1)

	homepageHandler := handler.HomepageHandler{
		CacheService:     mockCacheService,
		DashboardRepos:   mockDashboardRepos,
		DashboardService: mockDashboardService,
		HasDashboardData: true,
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
