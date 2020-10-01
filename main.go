package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lobsterdore/ops-dash/config"
	"github.com/lobsterdore/ops-dash/handler"
	"github.com/lobsterdore/ops-dash/service"

	"github.com/markbates/pkger"
)

func NewRouter(ctx context.Context, config config.Config) *http.ServeMux {
	router := http.NewServeMux()

	dashboardService := service.NewDashboardService(ctx, config)
	dashboardRepos, err := dashboardService.GetDashboardRepos(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}

	homepageHandler := handler.HomepageHandler{
		DashboardRepos:   dashboardRepos,
		DashboardService: dashboardService,
	}

	router.HandleFunc("/", homepageHandler.Http)

	return router
}

func main() {
	pkger.Include("/templates")

	log.Printf("Configuring server\n")
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("unable to retrieve configuration %s", err)
	}

	var runChan = make(chan os.Signal, 1)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.Server.Timeout.Server)*time.Second,
	)
	defer cancel()

	router := NewRouter(ctx, cfg)

	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.Timeout.Read) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.Timeout.Write) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.Timeout.Idle) * time.Second,
	}

	signal.Notify(runChan, os.Interrupt, syscall.SIGTSTP)

	log.Printf("Server is starting on %s\n", server.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
			} else {
				log.Fatalf("Server failed to start due to err: %v", err)
			}
		}
	}()

	interrupt := <-runChan
	log.Printf("Server is shutting down due to %+v\n", interrupt)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server was unable to gracefully shutdown due to err: %+v", err)
	}
}
