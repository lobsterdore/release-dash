package scm

import (
	"context"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=-mod=mod --source=scm.go --destination=../mocks/scm/scm.go
type ScmAdapter interface {
	GetChangelogForBranches(ctx context.Context, owner string, repo string, fromBranch string, toBranch string) (*[]ScmCommit, error)
	GetChangelogForTags(ctx context.Context, owner string, repo string, fromTag string, toTag string) (*[]ScmCommit, error)
	GetRepoBranch(ctx context.Context, owner string, repo string, branchName string) (*ScmRef, error)
	GetRepoFile(ctx context.Context, owner string, repo string, sha string, filePath string) ([]byte, error)
	GetUserRepos(ctx context.Context, user string) ([]ScmRepository, error)
}

type ScmCommit struct {
	AuthorAvatarUrl string
	Message         string
	HtmlUrl         string
}

type ScmRef struct {
	CurrentHash string
	Name        string
}

type ScmRepository struct {
	DefaultBranch string
	HtmlUrl       string
	Name          string
	OwnerName     string
}
