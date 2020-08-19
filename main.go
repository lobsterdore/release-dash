package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

type dashpageData struct {
	Commits []*github.RepositoryCommit
}

func getChangelog(ctx context.Context) *github.CommitsComparison {
	ghPat := os.Getenv("GH_PAT")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghPat},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	refFrom, _, err := client.Git.GetRef(ctx, "lobsterdore", "lobstercms", "tags/v2.1.0")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	refTo, _, err := client.Git.GetRef(ctx, "lobsterdore", "lobstercms", "tags/v2.7.0")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	comparison, _, err := client.Repositories.CompareCommits(ctx, "lobsterdore", "lobstercms", *refFrom.Object.SHA, *refTo.Object.SHA)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return comparison
}

func main() {

	tmpl := template.Must(template.ParseFiles("layout.html"))
	http.HandleFunc("/", func(respWriter http.ResponseWriter, request *http.Request) {
		comparison := getChangelog(request.Context())
		data := dashpageData{
			Commits: comparison.Commits,
		}

		tmpl.Execute(respWriter, data)
	})
	http.ListenAndServe(":8080", nil)

}
