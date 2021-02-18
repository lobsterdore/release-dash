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
		FromRef: "stg",
		ToRef:   "dev",
	}
	mockRepoChangelog := dashboard.DashboardRepoChangelog{
		ChangelogCommits: []dashboard.DashboardChangelogCommits{mockChangelogCommits},
		Repository:       mockRepo,
	}

	var mockRepoChangelogs []dashboard.DashboardRepoChangelog
	mockRepoChangelogs = append(mockRepoChangelogs, mockRepoChangelog)

	mockCacheService.
		EXPECT().
		Get("homepage_changelog_data").
		Times(1).
		Return(mockRepoChangelogs, true)

	homepageHandler := handler.HomepageHandler{
		CacheService:     mockCacheService,
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
	assert.Contains(t, resBody, mockRepoName)
	assert.Contains(t, resBody, "dev > stg")
	assert.Contains(t, resBody, mockMessage)
}

func TestHomepageHasRepoNoChanges(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockCacheService := mock_cache.NewMockCacheAdapter(ctrl)
	mockDashboardService := mock_dashboard.NewMockDashboardProvider(ctrl)

	mockCtx := context.Background()
	var mockRepoChangelogs []dashboard.DashboardRepoChangelog

	mockCacheService.
		EXPECT().
		Get("homepage_changelog_data").
		Times(1).
		Return(mockRepoChangelogs, true)

	homepageHandler := handler.HomepageHandler{
		CacheService:     mockCacheService,
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
