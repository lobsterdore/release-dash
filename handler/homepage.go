package handler

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/lobsterdore/ops-dash/service"
	"github.com/markbates/pkger"
)

type homepageData struct {
	RepoChangelogs []service.DashboardRepoChangelog
}

type HomepageHandler struct {
	DashboardRepos   *[]service.DashboardRepo
	DashboardService service.DashboardService
}

func readTemplateFile(name string) string {
	templateHandle, pkgerError := pkger.Open("/templates/" + name)
	if pkgerError != nil {
		panic(pkgerError)
	}
	defer templateHandle.Close()

	bytes, readError := ioutil.ReadAll(templateHandle)
	if readError != nil {
		panic(readError)
	}

	return string(bytes)
}

func (h *HomepageHandler) Http(respWriter http.ResponseWriter, request *http.Request) {
	log.Printf("Requested - '/' ")
	ctx := request.Context()
	tmpl, err := template.New("homepage").Parse(readTemplateFile("html/homepage.html"))
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
