package service

import (
	"log"
	"net/http"
	"text/template"

	"github.com/google/go-github/v32/github"
)

type dashpageData struct {
	Commits []*github.RepositoryCommit
}

func NewRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/", func(respWriter http.ResponseWriter, request *http.Request) {
		log.Printf("Requested - '/' ")
		ctx := request.Context()
		tmpl := template.Must(template.ParseFiles("layout.html"))

		client := GetGithubClient(ctx)
		comparison, err := GetChangelog(ctx, client)
		if err != nil {
			log.Fatal(err)
		}

		data := dashpageData{
			Commits: comparison.Commits,
		}

		tmpl.Execute(respWriter, data)
	})

	return router
}
