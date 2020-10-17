package handler

import (
	"context"
	"html/template"
	"log"
	"net/http"

	"github.com/lobsterdore/release-dash/service"
)

type homepageData struct {
	RepoChangelogs []service.DashboardRepoChangelog
}

type HomepageHandler struct {
	DashboardRepos   []service.DashboardRepo
	DashboardService service.DashboardProvider
}

func (h *HomepageHandler) Initialise(ctx context.Context) {
	dashboardRepos, err := h.DashboardService.GetDashboardRepos(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	h.DashboardRepos = dashboardRepos
	log.Printf("Homepage - Dashboard data refreshed")
}

func (h *HomepageHandler) Http(respWriter http.ResponseWriter, request *http.Request) {
	log.Printf("Requested - '/' ")
	ctx := request.Context()
	tmpl, err := template.New("homepage").Parse(service.ReadTemplateFile("html/homepage.html"))
	if err != nil {
		log.Println(err)
		return
	}

	var data = homepageData{
		RepoChangelogs: h.DashboardService.GetDashboardChangelogs(ctx, h.DashboardRepos),
	}

	err = tmpl.Execute(respWriter, data)
	if err != nil {
		respWriter.WriteHeader(http.StatusInternalServerError)
		_, _ = respWriter.Write([]byte(err.Error()))
		return
	}
}
