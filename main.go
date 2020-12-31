package main

import (
	"context"
	"os"
	"time"

	"github.com/markbates/pkger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lobsterdore/release-dash/config"
	"github.com/lobsterdore/release-dash/scm"
	"github.com/lobsterdore/release-dash/web"
)

func main() {
	_ = pkger.Include("/web/templates")
	_ = pkger.Include("/web/static")

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve configuration")
		os.Exit(3)
	}

	zLevel, err := zerolog.ParseLevel(cfg.Logging.Level)
	if err == nil {
		zerolog.SetGlobalLevel(zLevel)
	} else {
		log.Error().Err(err).Msgf("Could not get and set log level %s, using default", cfg.Logging.Level)
	}

	log.Info().Msg("Configuring server")
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.Server.Timeout.Server)*time.Second,
	)
	defer cancel()

	githubAdapter, err := scm.NewGithubAdapter(ctx, cfg.Github.Pat, cfg.Github.UrlDefault, cfg.Github.UrlUpload)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to setup Github client")
		os.Exit(3)
	}

	web.NewWeb(cfg, ctx, githubAdapter).Run(ctx)
}
