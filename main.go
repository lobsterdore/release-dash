package main

import (
	"log"

	"github.com/markbates/pkger"

	"github.com/lobsterdore/release-dash/config"
	"github.com/lobsterdore/release-dash/web"
)

func main() {
	_ = pkger.Include("/web/templates")

	log.Printf("Configuring server\n")
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("unable to retrieve configuration %s", err)
	}

	web.NewWeb(cfg).Run()
}
