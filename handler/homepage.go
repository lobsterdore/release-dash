package handler

import (
	"html/template"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/lobsterdore/ops-dash/service"
)

type homepageData struct {
	RepoChangelogs []homepageRepoChangelog
}

type homepageRepoChangelog struct {
	CommitsStg []github.RepositoryCommit
	CommitsPrd []github.RepositoryCommit
	Repository github.Repository
}

type HomepageHandler struct {
	DashboardRepos *[]service.DashboardRepo
	GithubService  service.GithubService
}

func (h *HomepageHandler) Http(respWriter http.ResponseWriter, request *http.Request) {
	log.Printf("Requested - '/' ")
	ctx := request.Context()
	tmpl := template.Must(template.ParseFiles("templates/html/homepage.html"))

	var repoChangelogs []homepageRepoChangelog

	for _, dashboardRepo := range *h.DashboardRepos {
		org := *dashboardRepo.Repository.Owner.Login
		repo := *dashboardRepo.Repository.Name

		repoChangelog := homepageRepoChangelog{
			Repository: *dashboardRepo.Repository,
		}

		comparisonStg, err := h.GithubService.GetChangelog(ctx, org, repo, "container-stg", "container-dev")
		if err == nil {
			repoChangelog.CommitsStg = comparisonStg.Commits
		}
		comparisonPrd, err := h.GithubService.GetChangelog(ctx, org, repo, "container-stg", "container-prd")
		if err == nil {
			repoChangelog.CommitsPrd = comparisonPrd.Commits
		}
		repoChangelogs = append(repoChangelogs, repoChangelog)
	}

	var data = homepageData{
		RepoChangelogs: repoChangelogs,
	}

	err := tmpl.Execute(respWriter, data)
	if err != nil {
		respWriter.WriteHeader(http.StatusInternalServerError)
		_, _ = respWriter.Write([]byte(err.Error()))
		return
	}
}
