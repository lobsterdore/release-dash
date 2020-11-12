package scm_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/github"
	"github.com/lobsterdore/release-dash/scm"
	"github.com/stretchr/testify/assert"
)

// Adapted from the go-github repo tests
func setupClient() (client *github.Client, mux *http.ServeMux, serverURL string, teardown func()) {
	mux = http.NewServeMux()
	apiHandler := http.NewServeMux()
	apiHandler.Handle("/api-v3/", http.StripPrefix("/api-v3", mux))
	server := httptest.NewServer(apiHandler)

	client = github.NewClient(nil)
	url, _ := url.Parse(server.URL + "/api-v3/")
	client.BaseURL = url
	client.UploadURL = url

	return client, mux, server.URL, server.Close
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func TestChangelogHasChanges(t *testing.T) {
	client, mux, _, teardown := setupClient()
	defer teardown()

	owner := "o"
	repo := "r"
	fromTag := "from-tag"
	fromSha := "812b303948b570247b727aeb8c1b187336ad4256"
	toTag := "to-tag"
	toSha := "3e0f3d8c432ca2a03a3222fb55de63934338022f"

	mux.HandleFunc("/repos/"+owner+"/"+repo+"/git/refs/tags/"+fromTag, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		  {
		    "ref": "tags`+fromTag+`",
		    "url": "https://api.github.com/repos/`+owner+`/`+repo+`/git/refs/tags/`+fromTag+`",
		    "object": {
		      "type": "commit",
		      "sha": "`+fromSha+`",
		      "url": "https://api.github.com/repos/o/r/git/commits/812b303948b570247b727aeb8c1b187336ad4256"
		    }
		  }`)
	})

	mux.HandleFunc("/repos/"+owner+"/"+repo+"/git/refs/tags/"+toTag, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		  {
		    "ref": "tags`+toTag+`",
		    "url": "https://api.github.com/repos/`+owner+`/`+repo+`/git/refs/tags/`+toTag+`",
		    "object": {
		      "type": "commit",
		      "sha": "`+toSha+`",
		      "url": "https://api.github.com/repos/o/r/git/commits/3e0f3d8c432ca2a03a3222fb55de63934338022f"
		    }
		  }`)
	})

	mux.HandleFunc("/repos/"+owner+"/"+repo+"/compare/"+fromSha+"..."+toSha, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `{
		  "base_commit": {
		    "sha": "s",
		    "commit": {
		      "author": { "name": "n" },
		      "committer": { "name": "n" },
			  "message": "m",
		      "tree": { "sha": "t" }
		    },
		    "author": { "login": "l" },
		    "committer": { "login": "l" },
		    "parents": [ { "sha": "s" } ]
		  },
		  "status": "s",
		  "ahead_by": 1,
		  "behind_by": 2,
		  "total_commits": 1,
		  "commits": [
		    {
			  "sha": "s",
			  "html_url": "h",
		      "commit": { "author": { "name": "n" }, "message": "m" },
		      "author": { "login": "l", "avatar_url": "a" },
		      "committer": { "login": "l" },
		      "parents": [ { "sha": "s" } ]
			},
		    {
				"sha": "s",
				"html_url": "h",
				"commit": { "author": { "name": "n" }, "message": "m" },
				"author": { "login": "l", "avatar_url": "a" },
				"committer": { "login": "l" },
				"parents": [ { "sha": "s" } ]
			}
		  ],
		  "files": [ { "filename": "f" } ],
		  "html_url":      "https://github.com/o/r/compare/b...h",
		  "permalink_url": "https://github.com/o/r/compare/o:bbcd538c8e72b8c175046e27cc8f907076331401...o:0328041d1152db8ae77652d1618a02e57f745f17",
		  "diff_url":      "https://github.com/o/r/compare/b...h.diff",
		  "patch_url":     "https://github.com/o/r/compare/b...h.patch",
		  "url":           "https://api.github.com/repos/o/r/compare/b...h"
		}`)
	})

	githubAdapter := scm.GithubAdapter{
		Client: client,
	}

	ctx := context.Background()

	changelog, err := githubAdapter.GetChangelog(ctx, owner, repo, fromTag, toTag)

	expectedChangelog := []scm.ScmCommit{
		scm.ScmCommit{
			AuthorAvatarUrl: "a",
			Message:         "m",
			HtmlUrl:         "h",
		},
		scm.ScmCommit{
			AuthorAvatarUrl: "a",
			Message:         "m",
			HtmlUrl:         "h",
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, &expectedChangelog, changelog)
}

func TestChangelogHasChangesMissingFromTag(t *testing.T) {
	client, mux, _, teardown := setupClient()
	defer teardown()

	owner := "o"
	repo := "r"
	fromTag := "from-tag"
	baseSha := "812b303948b570247b727aeb8c1b187336ad4256"
	toTag := "to-tag"
	toSha := "3e0f3d8c432ca2a03a3222fb55de63934338022f"

	mux.HandleFunc("/repos/"+owner+"/"+repo+"/git/refs/tags/"+fromTag, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/repos/"+owner+"/"+repo+"/git/refs/tags/"+toTag, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		  {
		    "ref": "tags`+toTag+`",
		    "url": "https://api.github.com/repos/`+owner+`/`+repo+`/git/refs/tags/`+toTag+`",
		    "object": {
		      "type": "commit",
		      "sha": "`+toSha+`",
		      "url": "https://api.github.com/repos/o/r/git/commits/3e0f3d8c432ca2a03a3222fb55de63934338022f"
		    }
		  }`)
	})

	mux.HandleFunc("/repos/"+owner+"/"+repo+"/compare/"+baseSha+"..."+toSha, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `{
		  "base_commit": {
		    "sha": "s",
		    "commit": {
		      "author": { "name": "n" },
		      "committer": { "name": "n" },
			  "message": "m",
		      "tree": { "sha": "t" }
		    },
		    "author": { "login": "l" },
		    "committer": { "login": "l" },
		    "parents": [ { "sha": "s" } ]
		  },
		  "status": "s",
		  "ahead_by": 1,
		  "behind_by": 2,
		  "total_commits": 1,
		  "commits": [
		    {
			  "sha": "s",
			  "html_url": "h",
		      "commit": { "author": { "name": "n" }, "message": "m" },
		      "author": { "login": "l", "avatar_url": "a" },
		      "committer": { "login": "l" },
		      "parents": [ { "sha": "s" } ]
		    }
		  ],
		  "files": [ { "filename": "f" } ],
		  "html_url":      "https://github.com/o/r/compare/b...h",
		  "permalink_url": "https://github.com/o/r/compare/o:bbcd538c8e72b8c175046e27cc8f907076331401...o:0328041d1152db8ae77652d1618a02e57f745f17",
		  "diff_url":      "https://github.com/o/r/compare/b...h.diff",
		  "patch_url":     "https://github.com/o/r/compare/b...h.patch",
		  "url":           "https://api.github.com/repos/o/r/compare/b...h"
		}`)
	})

	githubAdapter := scm.GithubAdapter{
		Client: client,
	}

	ctx := context.Background()

	changelog, err := githubAdapter.GetChangelog(ctx, owner, repo, fromTag, toTag)

	mockCommit := scm.ScmCommit{
		AuthorAvatarUrl: "a",
		Message:         "m",
		HtmlUrl:         "h",
	}
	expectedChangelog := []scm.ScmCommit{mockCommit}

	assert.NoError(t, err)
	assert.Equal(t, &expectedChangelog, changelog)
}
