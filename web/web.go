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
	"github.com/lobsterdore/release-dash/handler"
	"github.com/lobsterdore/release-dash/service"
)

type WebProvider interface {
	Run(ctx context.Context)
	SetupRouter(ctx context.Context) *http.ServeMux
}

type web struct {
	Config           config.Config
	DashboardService service.DashboardProvider
}

func NewWeb(cfg config.Config, dashboardService service.DashboardProvider) WebProvider {
	web := web{
		Config:           cfg,
		DashboardService: dashboardService,
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

	interrupt := <-runChan
	log.Printf("Server is shutting down due to %+v\n", interrupt)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server was unable to gracefully shutdown due to err: %+v", err)
	}
}

func (w web) SetupRouter(ctx context.Context) *http.ServeMux {
	router := http.NewServeMux()

	dashboardRepos, err := w.DashboardService.GetDashboardRepos(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}

	homepageHandler := handler.HomepageHandler{
		DashboardRepos:   dashboardRepos,
		DashboardService: w.DashboardService,
	}

	router.HandleFunc("/", homepageHandler.Http)

	return router
}
