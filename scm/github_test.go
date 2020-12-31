package scm_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/lobsterdore/release-dash/scm"
	"github.com/lobsterdore/release-dash/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestChangelogHasChanges(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	owner := "o"
	repo := "test-repo"
	fromTag := "from-tag"
	toTag := "to-tag"

	githubAdapter := scm.GithubAdapter{
		Client: client,
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
		Client: client,
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
		Client: client,
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
	repo := "test-repo"
	owner := "o"

	githubAdapter := scm.GithubAdapter{
		Client: client,
	}

	ctx := context.Background()

	expectedScmRepos := []scm.ScmRepository{scm.ScmRepository{
		DefaultBranch: defaultBranch,
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
		Client: client,
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
		Client: client,
	}

	ctx := context.Background()

	expectedScmBranch := scm.ScmBranch{
		CurrentHash: sha,
		Name:        branch,
	}

	scmBranch, err := githubAdapter.GetRepoBranch(ctx, owner, repo, branch)

	assert.NoError(t, err)
	assert.Equal(t, &expectedScmBranch, scmBranch)
}

func TestGetRepoBranchError(t *testing.T) {
	client, teardown := testsupport.SetupGithubClientMock()
	defer teardown()

	branch := "main"
	repo := "500"
	owner := "o"

	githubAdapter := scm.GithubAdapter{
		Client: client,
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
		Client: client,
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
		Client: client,
	}

	ctx := context.Background()

	repoFile, err := githubAdapter.GetRepoFile(ctx, owner, repo, sha, path)

	assert.NoError(t, err)
	assert.Nil(t, repoFile)
}
