package scm

import (
	"context"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=-mod=mod --source=scm.go --destination=../mocks/scm/scm.go
type ScmAdaptor interface {
	GetChangelog(ctx context.Context, owner string, repo string, fromTag string, toTag string) (*[]ScmCommit, error)
	GetRepoBranch(ctx context.Context, owner string, repo string, branchName string) (*ScmBranch, error)
	GetRepoFile(ctx context.Context, owner string, repo string, sha string, filePath string) ([]byte, error)
	GetUserRepos(ctx context.Context, user string) ([]ScmRepository, error)
}

type ScmBranch struct {
	CurrentHash string
	Name        string
}

type ScmCommit struct {
	AuthorAvatarUrl string
	Message         string
	HtmlUrl         string
}

type ScmRepository struct {
	Name      string
	OwnerName string
}
