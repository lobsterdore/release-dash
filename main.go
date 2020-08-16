package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()

	ghPat := os.Getenv("GH_PAT")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghPat},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repo, _, err := client.Repositories.Get(ctx, "lobsterdore", "lobstercms")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	releases, _, err := client.Repositories.ListReleases(ctx, "lobsterdore", "lobstercms", nil)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(repo.Name)
	fmt.Println(releases)
}
