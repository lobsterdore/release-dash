package web

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
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
	Config             config.Config
	DashboardService   dashboard.DashboardProvider
	HealthcheckHandler *handler.HealthcheckHandler
	HomepageHandler    *handler.HomepageHandler
}

func NewWeb(cfg config.Config, ctx context.Context, scmService scm.ScmAdapter, cacheService cache.CacheAdapter) WebProvider {
	dashboardService := dashboard.NewDashboardService(ctx, cfg, scmService)

	healthcheckHandler := handler.NewHealthcheckHandler()
	homepageHandler := handler.NewHomepageHandler(dashboardService, cacheService)

	web := web{
		Config:             cfg,
		DashboardService:   dashboardService,
		HealthcheckHandler: healthcheckHandler,
		HomepageHandler:    homepageHandler,
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
		IdleTimeout:  time.Duration(w.Config.Server.Timeout.Idle) * time.Second,
		ReadTimeout:  time.Duration(w.Config.Server.Timeout.Read) * time.Second,
		WriteTimeout: time.Duration(w.Config.Server.Timeout.Write) * time.Second,
	}

	_ = notifySignals(runChan)

	log.Log().Msgf("Server is starting on %s", server.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
			} else {
				log.Fatal().Err(err).Msg("Server failed to start")
			}
		}
	}()

	w.HomepageHandler.FetchReposTicker(w.Config.Github.RepoFetchTimerSeconds)
	w.HomepageHandler.FetchChangelogsTicker(w.Config.Github.ChangelogFetchTimerSeconds)

	interrupt := <-runChan
	log.Log().Msgf("Server is shutting down due to %+v", interrupt)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server was unable to gracefully shutdown")
	}
}

func (w web) SetupRouter(ctx context.Context) *http.ServeMux {
	router := http.NewServeMux()
	fs := http.FileServer(pkger.Dir("/web/static"))

	router.Handle("/static/", http.StripPrefix("/static", fs))
	router.HandleFunc("/", w.HomepageHandler.Http)
	router.HandleFunc("/healthcheck", w.HealthcheckHandler.Http)

	if w.Config.Profiling.Enabled {
		log.Log().Msg("Enabling profiling")
		router.HandleFunc("/debug/pprof/", pprof.Index)
	}
	return router
}
