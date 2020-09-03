package handler

import (
	"html/template"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/lobsterdore/ops-dash/service"
)

type dashpageData struct {
	Commits []github.RepositoryCommit
}

type HttpHandler struct {
	GithubService service.GithubService
}

func (h *HttpHandler) Homepage(respWriter http.ResponseWriter, request *http.Request) {
	log.Printf("Requested - '/' ")
	ctx := request.Context()
	tmpl := template.Must(template.ParseFiles("html/homepage.html"))

	comparison, err := h.GithubService.GetChangelog(ctx, "lobsterdore", "lobstercms", "dev", "prd")
	if err != nil {
		log.Fatal(err)
	}

	data := dashpageData{
		Commits: comparison.Commits,
	}

	err = tmpl.Execute(respWriter, data)
	if err != nil {
		log.Fatal(err)
	}
}
