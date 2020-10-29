package main

import (
	"context"
	"log"
	"time"

	"github.com/markbates/pkger"

	"github.com/lobsterdore/release-dash/config"
	"github.com/lobsterdore/release-dash/web"
)

func main() {
	_ = pkger.Include("/web/templates")
	_ = pkger.Include("/web/static")

	log.Printf("Configuring server\n")
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("unable to retrieve configuration %s", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.Server.Timeout.Server)*time.Second,
	)
	defer cancel()

	web.NewWeb(cfg, ctx).Run(ctx)
}
