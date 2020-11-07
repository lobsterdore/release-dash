package web

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lobsterdore/release-dash/cache"
	"github.com/lobsterdore/release-dash/config"
	"github.com/lobsterdore/release-dash/dashboard"
	"github.com/lobsterdore/release-dash/logging"
	"github.com/lobsterdore/release-dash/scm"
	"github.com/lobsterdore/release-dash/web/handler"

	"github.com/markbates/pkger"
	"github.com/rs/zerolog/log"

	accesslog "github.com/mash/go-accesslog"
)

type WebProvider interface {
	Run(ctx context.Context)
	SetupRouter(ctx context.Context) *http.ServeMux
}

type web struct {
	Config           config.Config
	DashboardService dashboard.DashboardProvider
	HomepageHandler  *handler.HomepageHandler
}

func NewWeb(cfg config.Config, ctx context.Context, cacheService cache.CacheAdaptor, scmService scm.ScmAdaptor) WebProvider {
	var placeholderRepos []dashboard.DashboardRepo
	dashboardService := dashboard.NewDashboardService(ctx, cfg, scmService)

	homepageHandler := handler.HomepageHandler{
		CacheService:     cacheService,
		DashboardRepos:   placeholderRepos,
		DashboardService: dashboardService,
		HasDashboardData: false,
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
	httpLogger := logging.HttpLogger{}

	server := &http.Server{
		Addr:         w.Config.Server.Host + ":" + w.Config.Server.Port,
		Handler:      accesslog.NewLoggingHandler(router, httpLogger),
		ReadTimeout:  time.Duration(w.Config.Server.Timeout.Read) * time.Second,
		WriteTimeout: time.Duration(w.Config.Server.Timeout.Write) * time.Second,
		IdleTimeout:  time.Duration(w.Config.Server.Timeout.Idle) * time.Second,
	}

	signal.Notify(runChan, os.Interrupt, syscall.SIGTSTP)

	log.Printf("Server is starting on %s", server.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
			} else {
				log.Fatal().Err(err).Msg("Server failed to start")
			}
		}
	}()

	w.HomepageHandler.FetchReposTicker(w.Config.Github.FetchTimerSeconds)

	interrupt := <-runChan
	log.Printf("Server is shutting down due to %+v", interrupt)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server was unable to gracefully shutdown")
	}
}

func (w web) SetupRouter(ctx context.Context) *http.ServeMux {
	router := http.NewServeMux()
	fs := http.FileServer(pkger.Dir("/web/static"))

	router.Handle("/static/", http.StripPrefix("/static", fs))
	router.HandleFunc("/", w.HomepageHandler.Http)

	return router
}
