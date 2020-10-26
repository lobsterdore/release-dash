package handler

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/lobsterdore/release-dash/cache"
	"github.com/lobsterdore/release-dash/service"
)

type HomepageData struct {
	RepoChangelogs []service.DashboardRepoChangelog
}

type HomepageHandler struct {
	CacheService     cache.CacheProvider
	DashboardRepos   []service.DashboardRepo
	DashboardService service.DashboardProvider
	HasDashboardData bool
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
	log.Printf("Homepage - Dashboard data fetching")
	dashboardRepos, err := h.DashboardService.GetDashboardRepos(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	h.DashboardRepos = dashboardRepos
	h.HasDashboardData = true
	log.Printf("Homepage - Dashboard data refreshed")
}

func (h *HomepageHandler) Http(respWriter http.ResponseWriter, request *http.Request) {
	log.Printf("Requested - '/' ")
	ctx := request.Context()

	var tmpl *template.Template
	var data HomepageData
	var err error

	if h.HasDashboardData {
		tmpl, err = template.New("homepage").Parse(service.ReadTemplateFile("html/homepage.html"))
		if err != nil {
			log.Println(err)
			return
		}
		cachedData, found := h.CacheService.Get("homepage_data")
		if found {
			data = cachedData.(HomepageData)
		} else {
			data = HomepageData{
				RepoChangelogs: h.DashboardService.GetDashboardChangelogs(ctx, h.DashboardRepos),
			}
			h.CacheService.Set("homepage_data", data)
		}
	} else {
		tmpl, err = template.New("homepage_loading").Parse(service.ReadTemplateFile("html/homepage_loading.html"))
		if err != nil {
			log.Println(err)
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
