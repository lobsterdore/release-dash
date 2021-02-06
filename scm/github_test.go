package scm_test

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/flowchartsman/retry"
	"github.com/google/go-github/github"
	"github.com/lobsterdore/release-dash/scm"
	"github.com/lobsterdore/release-dash/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestCheckForRetryRateLimited(t *testing.T) {
	resp := &github.Response{
		Response: &http.Response{
			StatusCode: 403,
		},
	}
	err := &github.RateLimitError{}
	retryErr := scm.CheckForRetry(resp, err)

	assert.Error(t, retryErr)
	assert.NotEqual(t, retryErr, err)
}

func TestCheckForRetryNotRateLimited(t *testing.T) {
	resp := &github.Response{
		Response: &http.Response{
			StatusCode: 500,
		},
	}
	err := errors.New("Not rate limited")
	retryErr := scm.CheckForRetry(resp, err)

	assert.Error(t, retryErr)
	assert.Equal(t, retryErr.Error(), err.Error())
}

func TestCheckForRetryNoError(t *testing.T) {
	resp := &github.Response{
		Response: &http.Response{
			StatusCode: 200,
		},
	}
	retryErr := scm.CheckForRetry(resp, nil)

	assert.NoError(t, retryErr)
}

func TestChangelogHasChanges(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	owner := "o"
	repo := "test-repo"
	fromTag := "from-tag"
	toTag := "to-tag"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	changelog, err := githubAdapter.GetChangelog(ctx, owner, repo, fromTag, toTag)

	expectedChangelog := []scm.ScmCommit{
		scm.ScmCommit{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
		scm.ScmCommit{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, &expectedChangelog, changelog)
}

func TestChangelogHasChangesMissingFromTag(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	owner := "o"
	repo := "test-repo"
	fromTag := "missing-tag"
	toTag := "to-tag"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	changelog, err := githubAdapter.GetChangelog(ctx, owner, repo, fromTag, toTag)

	expectedChangelog := []scm.ScmCommit{
		scm.ScmCommit{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
		scm.ScmCommit{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, &expectedChangelog, changelog)
}

func TestChangelogHasChangesMissingToTag(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	owner := "o"
	repo := "test-repo"
	fromTag := "missing-tag"
	toTag := "missing-tag"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	changelog, err := githubAdapter.GetChangelog(ctx, owner, repo, fromTag, toTag)

	assert.NoError(t, err)
	assert.Nil(t, changelog)
}

func TestUserReposHasRepos(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	defaultBranch := "main"
	htmlUrl := "url"
	repo := "test-repo"
	owner := "o"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	expectedScmRepos := []scm.ScmRepository{scm.ScmRepository{
		DefaultBranch: defaultBranch,
		HtmlUrl:       htmlUrl,
		Name:          repo,
		OwnerName:     owner,
	}}

	scmRepos, err := githubAdapter.GetUserRepos(ctx, "")

	assert.NoError(t, err)
	assert.Equal(t, expectedScmRepos, scmRepos)
}

func TestUserReposListError(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	scmRepos, err := githubAdapter.GetUserRepos(ctx, "500")

	assert.Error(t, err)
	assert.Nil(t, scmRepos)
}

func TestGetRepoBranchHasBranch(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	branch := "main"
	repo := "test-repo"
	owner := "o"
	sha := "s"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	expectedScmBranch := scm.ScmRef{
		CurrentHash: sha,
		Name:        branch,
	}

	ScmRef, err := githubAdapter.GetRepoBranch(ctx, owner, repo, branch)

	assert.NoError(t, err)
	assert.Equal(t, &expectedScmBranch, ScmRef)
}

func TestGetRepoBranchError(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	branch := "main"
	repo := "500"
	owner := "o"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	_, err := githubAdapter.GetRepoBranch(ctx, owner, repo, branch)

	assert.Error(t, err)
}

func TestGetRepoFileHasFile(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	repo := "test-repo"
	owner := "o"
	sha := "s"
	path := ".releasedash.yml"
	content := "LS0tCgplbnZpcm9ubWVudF90YWdzOgogIC0gZnJvbS10YWcKICAtIHRvLXRhZwpuYW1lOiByCg=="

	expectedRepoFile, _ := base64.StdEncoding.DecodeString(content)

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	repoFile, err := githubAdapter.GetRepoFile(ctx, owner, repo, sha, path)

	assert.NoError(t, err)
	assert.Equal(t, expectedRepoFile, repoFile)
}

func TestGetRepoFileMissingFile(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	repo := "missingfile"
	owner := "o"
	sha := "s"
	path := ".releasedash.yml"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	repoFile, err := githubAdapter.GetRepoFile(ctx, owner, repo, sha, path)

	assert.NoError(t, err)
	assert.Nil(t, repoFile)
}
