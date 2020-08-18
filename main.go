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

	refFrom, _, err := client.Git.GetRef(ctx, "lobsterdore", "lobstercms", "tags/v2.1.0")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	refTo, _, err := client.Git.GetRef(ctx, "lobsterdore", "lobstercms", "tags/v2.7.0")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	comparison, _, err := client.Repositories.CompareCommits(ctx, "lobsterdore", "lobstercms", *refFrom.Object.SHA, *refTo.Object.SHA)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for i := 0; i < len(comparison.Commits); i++ {
		fmt.Println(*comparison.Commits[i].Commit.Message)
	}

}
