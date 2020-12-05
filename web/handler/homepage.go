package handler

import (
	"context"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lobsterdore/release-dash/asset"
	"github.com/lobsterdore/release-dash/dashboard"
	"github.com/lobsterdore/release-dash/web/templatefns"
)

type HomepageData struct {
	RepoChangelogs []dashboard.DashboardRepoChangelog
}

type HomepageHandler struct {
	DashboardRepos   []dashboard.DashboardRepo
	DashboardService dashboard.DashboardProvider
	HasChangelogData bool
	HasDashboardData bool
	RepoChangelogs   []dashboard.DashboardRepoChangelog
}

func NewHomepage(dashboardService *dashboard.DashboardService) *HomepageHandler {
	var placeholderRepos []dashboard.DashboardRepo
	var placeholderChangelogs []dashboard.DashboardRepoChangelog

	homepageHandler := HomepageHandler{
		DashboardRepos:   placeholderRepos,
		DashboardService: dashboardService,
		HasChangelogData: false,
		HasDashboardData: false,
		RepoChangelogs:   placeholderChangelogs,
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
				h.FetchRepos(ctx)
			}()
		}
	}()
}

func (h *HomepageHandler) FetchRepos(ctx context.Context) {
	log.Print("Dashboard repo data fetching")
	dashboardRepos, err := h.DashboardService.GetDashboardRepos(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Dashboard repo data fetch failed")
		return
	}
	h.DashboardRepos = dashboardRepos
	h.HasDashboardData = true
	log.Print("Dashboard repo data refreshed")
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
				h.FetchChangelogs(ctx)
			}()
		}
	}()
}

func (h *HomepageHandler) FetchChangelogs(ctx context.Context) {
	log.Print("Dashboard changelog data fetching")
	if h.HasDashboardData {
		dashboardChangelogs := h.DashboardService.GetDashboardChangelogs(ctx, h.DashboardRepos)
		h.RepoChangelogs = dashboardChangelogs
		h.HasChangelogData = true
		log.Print("Dashboard changelog repo data refreshed")
	} else {
		log.Print("Dashboard repo data not present yet")
	}
}

func (h *HomepageHandler) Http(respWriter http.ResponseWriter, request *http.Request) {
	var tmpl *template.Template
	var data HomepageData
	var err error

	if h.HasDashboardData && h.HasChangelogData {
		tmpl, err = template.New("homepage").Funcs(templatefns.TemplateFnsMap).Parse(asset.ReadTemplateFile("html/base.html"))
		if err != nil {
			log.Print(err)
			return
		}

		tmpl, err = tmpl.Parse(asset.ReadTemplateFile("html/homepage.html"))
		if err != nil {
			log.Print(err)
			return
		}
		data = HomepageData{
			RepoChangelogs: h.RepoChangelogs,
		}
	} else {
		tmpl, err = template.New("homepage_loading").Funcs(templatefns.TemplateFnsMap).Parse(asset.ReadTemplateFile("html/base.html"))
		if err != nil {
			log.Print(err)
			return
		}

		tmpl, err = tmpl.Parse(asset.ReadTemplateFile("html/homepage_loading.html"))
		if err != nil {
			log.Print(err)
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
