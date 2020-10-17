package web

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lobsterdore/release-dash/config"
	"github.com/lobsterdore/release-dash/service"
	"github.com/lobsterdore/release-dash/web/handler"
)

type WebProvider interface {
	Run(ctx context.Context)
	SetupRouter(ctx context.Context) *http.ServeMux
}

type web struct {
	Config           config.Config
	DashboardService service.DashboardProvider
	HomepageHandler  *handler.HomepageHandler
}

func NewWeb(cfg config.Config, dashboardService service.DashboardProvider) WebProvider {
	var placeholderRepos []service.DashboardRepo

	homepageHandler := handler.HomepageHandler{
		DashboardRepos:   placeholderRepos,
		DashboardService: dashboardService,
	}

	web := web{
		Config:           cfg,
		DashboardService: dashboardService,
		HomepageHandler:  &homepageHandler,
	}
	return web
}

func (w web) Run(ctx context.Context) {
	var runChan = make(chan os.Signal, 1)

	router := w.SetupRouter(ctx)

	server := &http.Server{
		Addr:         w.Config.Server.Host + ":" + w.Config.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(w.Config.Server.Timeout.Read) * time.Second,
		WriteTimeout: time.Duration(w.Config.Server.Timeout.Write) * time.Second,
		IdleTimeout:  time.Duration(w.Config.Server.Timeout.Idle) * time.Second,
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

	w.HomepageHandler.Initialise(ctx)

	interrupt := <-runChan
	log.Printf("Server is shutting down due to %+v\n", interrupt)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server was unable to gracefully shutdown due to err: %+v", err)
	}
}

func (w web) SetupRouter(ctx context.Context) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/", w.HomepageHandler.Http)

	return router
}
