package handler

import (
	"html/template"
	"log"
	"net/http"

	"github.com/lobsterdore/ops-dash/service"
)

type homepageData struct {
	RepoChangelogs []service.DashboardRepoChangelog
}

type HomepageHandler struct {
	DashboardRepos   *[]service.DashboardRepo
	DashboardService service.DashboardService
}

func (h *HomepageHandler) Http(respWriter http.ResponseWriter, request *http.Request) {
	log.Printf("Requested - '/' ")
	ctx := request.Context()
	tmpl := template.Must(template.ParseFiles("templates/html/homepage.html"))

	var data = homepageData{
		RepoChangelogs: h.DashboardService.GetDashboardChangelogs(ctx, h.DashboardRepos),
	}

	err := tmpl.Execute(respWriter, data)
	if err != nil {
		respWriter.WriteHeader(http.StatusInternalServerError)
		_, _ = respWriter.Write([]byte(err.Error()))
		return
	}
}
