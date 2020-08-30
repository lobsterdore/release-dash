package handler

import (
	"html/template"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/lobsterdore/ops-dash/service"
)

type HttpHandler struct {
	GithubService *service.GithubService
}

type dashpageData struct {
	Commits []*github.RepositoryCommit
}

func (h *HttpHandler) Homepage(respWriter http.ResponseWriter, request *http.Request) {
	log.Printf("Requested - '/' ")
	ctx := request.Context()
	tmpl := template.Must(template.ParseFiles("layout.html"))

	comparison, err := h.GithubService.GetChangelog(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	data := dashpageData{
		Commits: comparison.Commits,
	}

	tmpl.Execute(respWriter, data)
}
