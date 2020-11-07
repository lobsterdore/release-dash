package main

import (
	"context"
	"os"
	"time"

	"github.com/markbates/pkger"
	"github.com/rs/zerolog/log"

	"github.com/lobsterdore/release-dash/cache"
	"github.com/lobsterdore/release-dash/config"
	"github.com/lobsterdore/release-dash/scm"
	"github.com/lobsterdore/release-dash/web"
)

func main() {
	_ = pkger.Include("/web/templates")
	_ = pkger.Include("/web/static")

	log.Print("Configuring server")
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve configuration")
		os.Exit(3)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.Server.Timeout.Server)*time.Second,
	)
	defer cancel()

	cacheService := cache.NewLocalCacheService(
		cfg.Cache.DefaultExpirationMinutes,
		cfg.Cache.CleanupIntervalMinutes,
	)

	githubService := scm.NewGithubService(ctx, cfg.Github.Pat)

	web.NewWeb(cfg, ctx, cacheService, githubService).Run(ctx)
}
