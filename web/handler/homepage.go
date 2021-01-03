package handler

import (
	"context"
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lobsterdore/release-dash/asset"
	"github.com/lobsterdore/release-dash/cache"
	"github.com/lobsterdore/release-dash/dashboard"
	"github.com/lobsterdore/release-dash/web/templatefns"
)

type HomepageData struct {
	RepoChangelogs []dashboard.DashboardRepoChangelog
}

type HomepageHandler struct {
	CacheService     cache.CacheAdapter
	DashboardService dashboard.DashboardProvider
}

func NewHomepageHandler(dashboardService *dashboard.DashboardService, cacheService cache.CacheAdapter) *HomepageHandler {
	homepageHandler := HomepageHandler{
		CacheService:     cacheService,
		DashboardService: dashboardService,
	}

	return &homepageHandler
}

func (h *HomepageHandler) FetchReposTicker(timerSeconds int) {
	mux := &sync.Mutex{}
	go func() {
		ctx := context.Background()
		ticker := time.NewTicker(time.Duration(timerSeconds) * time.Second)
		for ; true; <-ticker.C {
			func() {
				mux.Lock()
				defer mux.Unlock()
				expireSeconds := strconv.Itoa(timerSeconds * 2)
				h.FetchRepos(ctx, expireSeconds)
			}()
		}
	}()
}

func (h *HomepageHandler) FetchRepos(ctx context.Context, expireSeconds string) {
	log.Info().Msg("Dashboard repo data fetching")
	dashboardRepos, err := h.DashboardService.GetDashboardRepos(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Dashboard repo data fetch failed")
		return
	}
	h.CacheService.Set("homepage_repo_data", dashboardRepos, expireSeconds)
	log.Info().Msg("Dashboard repo data refreshed")
}

func (h *HomepageHandler) FetchChangelogsTicker(timerSeconds int) {
	mux := &sync.Mutex{}
	go func() {
		ctx := context.Background()
		ticker := time.NewTicker(time.Duration(timerSeconds) * time.Second)
		for ; true; <-ticker.C {
			func() {
				mux.Lock()
				defer mux.Unlock()
				expireSeconds := strconv.Itoa(timerSeconds * 2)
				h.FetchChangelogs(ctx, expireSeconds)
			}()
		}
	}()
}

func (h *HomepageHandler) FetchChangelogs(ctx context.Context, expireSeconds string) {
	log.Info().Msg("Dashboard changelog data fetching")

	cachedData, found := h.CacheService.Get("homepage_repo_data")
	if found {
		dashboardRepos := cachedData.([]dashboard.DashboardRepo)
		dashboardChangelogs := h.DashboardService.GetDashboardChangelogs(ctx, dashboardRepos)
		h.CacheService.Set("homepage_changelog_data", dashboardChangelogs, expireSeconds)
		log.Info().Msg("Dashboard changelog repo data refreshed")
	} else {
		log.Info().Msg("Dashboard repo data not present yet")
	}
}

func (h *HomepageHandler) Http(respWriter http.ResponseWriter, request *http.Request) {
	var tmpl *template.Template
	var data HomepageData
	var err error

	cachedData, found := h.CacheService.Get("homepage_changelog_data")
	if found {
		repoChangelogs := cachedData.([]dashboard.DashboardRepoChangelog)
		tmpl, err = template.New("homepage").Funcs(templatefns.TemplateFnsMap).Parse(asset.ReadTemplateFile("html/base.html"))
		if err != nil {
			log.Error().Err(err).Msg("Could not get html/base.html")
			return
		}

		tmpl, err = tmpl.Parse(asset.ReadTemplateFile("html/homepage.html"))
		if err != nil {
			log.Error().Err(err).Msg("Could not get html/homepage.html")
			return
		}
		data = HomepageData{
			RepoChangelogs: repoChangelogs,
		}
	} else {
		tmpl, err = template.New("homepage_loading").Funcs(templatefns.TemplateFnsMap).Parse(asset.ReadTemplateFile("html/base.html"))
		if err != nil {
			log.Error().Err(err).Msg("Could not get html/base.html")
			return
		}

		tmpl, err = tmpl.Parse(asset.ReadTemplateFile("html/homepage_loading.html"))
		if err != nil {
			log.Error().Err(err).Msg("Could not get html/homepage_loading.html")
			return
		}
	}

	err = tmpl.Execute(respWriter, data)
	if err != nil {
		respWriter.WriteHeader(http.StatusInternalServerError)
		_, _ = respWriter.Write([]byte(err.Error()))
		return
	}
}
