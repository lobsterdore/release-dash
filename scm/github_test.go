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

func TestGetChangelogForBranchesHasChanges(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	owner := "o"
	repo := "test-repo"
	fromBranch := "from-branch"
	toBranch := "to-branch"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	changelog, err := githubAdapter.GetChangelogForBranches(ctx, owner, repo, fromBranch, toBranch)

	expectedChangelog := []scm.ScmCommit{
		{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
		{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, &expectedChangelog, changelog)
}

func TestGetChangelogForTagsHasChanges(t *testing.T) {
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

	changelog, err := githubAdapter.GetChangelogForTags(ctx, owner, repo, fromTag, toTag)

	expectedChangelog := []scm.ScmCommit{
		{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
		{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, &expectedChangelog, changelog)
}

func TestGetChangelogForTagsHasChangesMissingFromTag(t *testing.T) {
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

	changelog, err := githubAdapter.GetChangelogForTags(ctx, owner, repo, fromTag, toTag)

	expectedChangelog := []scm.ScmCommit{
		{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
		{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, &expectedChangelog, changelog)
}

func TestGetChangelogForTagsHasChangesMissingToTag(t *testing.T) {
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

	changelog, err := githubAdapter.GetChangelogForTags(ctx, owner, repo, fromTag, toTag)

	assert.NoError(t, err)
	assert.Nil(t, changelog)
}

func TestGetRepoCommitsForShaHasCommits(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	owner := "o"
	repo := "test-repo"
	toSha := "3e0f3d8c432ca2a03a3222fb55de63934338022f"
	foundSha := "812b303948b570247b727aeb8c1b187336ad4256"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	commits, err := githubAdapter.GetRepoCommitsForSha(ctx, owner, repo, toSha)

	expectedCommits := []*github.RepositoryCommit{{
		SHA: &foundSha,
	}}

	assert.NoError(t, err)
	assert.Equal(t, expectedCommits, commits)

}

func TestGetRepoCommitsForShaError(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	owner := "o"
	repo := "500"
	toSha := "3e0f3d8c432ca2a03a3222fb55de63934338022f"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	commits, err := githubAdapter.GetRepoCommitsForSha(ctx, owner, repo, toSha)

	assert.Error(t, err)
	assert.Nil(t, commits)

}

func TestGetRepoCompareCommitsHasCommits(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	repo := "test-repo"
	owner := "o"
	fromSha := "812b303948b570247b727aeb8c1b187336ad4256"
	toSha := "3e0f3d8c432ca2a03a3222fb55de63934338022f"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	expectedComparison := []scm.ScmCommit{
		{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
		{
			AuthorAvatarUrl: "a",
			Message:         "test-commit",
			HtmlUrl:         "h",
		},
	}

	comparison, err := githubAdapter.GetRepoCompareCommits(ctx, owner, repo, fromSha, toSha)

	assert.NoError(t, err)
	assert.Equal(t, &expectedComparison, comparison)
}

func TestGetRepoCompareCommitsError(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	repo := "500"
	owner := "o"
	fromSha := "812b303948b570247b727aeb8c1b187336ad4256"
	toSha := "3e0f3d8c432ca2a03a3222fb55de63934338022f"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	comparison, err := githubAdapter.GetRepoCompareCommits(ctx, owner, repo, fromSha, toSha)

	assert.Error(t, err)
	assert.Nil(t, comparison)
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

func TestGetRepoTagHasTag(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	owner := "o"
	repo := "test-repo"
	tag := "from-tag"
	sha := "812b303948b570247b727aeb8c1b187336ad4256"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	scmTag, err := githubAdapter.GetRepoTag(ctx, owner, repo, tag)

	expectedScmTag := scm.ScmRef{
		CurrentHash: sha,
		Name:        tag,
	}

	assert.NoError(t, err)
	assert.Equal(t, &expectedScmTag, scmTag)
}

func TestGetRepoTagMissingTag(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	owner := "o"
	repo := "test-repo"
	tag := " missing-tag"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	scmTag, err := githubAdapter.GetRepoTag(ctx, owner, repo, tag)

	assert.NoError(t, err)
	assert.Nil(t, scmTag)
}

func TestGetRepoTagError(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	owner := "o"
	repo := "500"
	tag := "from-tag"

	githubAdapter := scm.GithubAdapter{
		Client:  client,
		Retrier: retry.NewRetrier(5, 5*time.Second, 30*time.Second),
	}

	ctx := context.Background()

	scmTag, err := githubAdapter.GetRepoTag(ctx, owner, repo, tag)

	assert.Error(t, err)
	assert.Nil(t, scmTag)
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

	expectedScmRepos := []scm.ScmRepository{{
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
