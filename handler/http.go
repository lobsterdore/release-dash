package handler

import (
	"html/template"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/lobsterdore/ops-dash/service"
)

type dashpageData struct {
	CommitsStg []github.RepositoryCommit
	CommitsPrd []github.RepositoryCommit
}

type HttpHandler struct {
	GithubService service.GithubService
}

func (h *HttpHandler) Homepage(respWriter http.ResponseWriter, request *http.Request) {
	log.Printf("Requested - '/' ")
	ctx := request.Context()
	tmpl := template.Must(template.ParseFiles("html/homepage.html"))

	comparisonStg, err := h.GithubService.GetChangelog(ctx, "JSainsburyPLC", "smartshop-api-go-canary", "container-stg", "container-dev")
	if err != nil {
		respWriter.WriteHeader(http.StatusInternalServerError)
		_, _ = respWriter.Write([]byte(err.Error()))
		return
	}
	comparisonPrd, err := h.GithubService.GetChangelog(ctx, "JSainsburyPLC", "smartshop-api-go-canary", "container-stg", "container-prd")
	if err != nil {
		respWriter.WriteHeader(http.StatusInternalServerError)
		_, _ = respWriter.Write([]byte(err.Error()))
		return
	}

	data := dashpageData{
		CommitsStg: comparisonStg.Commits,
		CommitsPrd: comparisonPrd.Commits,
	}

	err = tmpl.Execute(respWriter, data)
	if err != nil {
		respWriter.WriteHeader(http.StatusInternalServerError)
		_, _ = respWriter.Write([]byte(err.Error()))
		return
	}
}
