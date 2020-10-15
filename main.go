package main

import (
	"context"
	"log"
	"time"

	"github.com/markbates/pkger"

	"github.com/lobsterdore/release-dash/config"
	"github.com/lobsterdore/release-dash/service"
	"github.com/lobsterdore/release-dash/web"
)

func main() {
	_ = pkger.Include("/web/templates")

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

	dashboardService := service.NewDashboardService(ctx, cfg)

	web.NewWeb(cfg, dashboardService).Run(ctx)
}
