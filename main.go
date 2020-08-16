package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()

	gh_pat := os.Getenv("GH_PAT")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gh_pat},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repo, _, err := client.Repositories.Get(ctx, "JSainsburyPLC", "api-issa-canary")

	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			log.Println("hit rate limit")
		}

		fmt.Println("Bang")
		os.Exit(1)
	}

	fmt.Println(repo)
}
