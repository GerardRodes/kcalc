package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	_ "expvar"
)

var argHTTPPort = flag.String("http-port", "8080", "")

func Serve(ctx context.Context) error {
	router := httprouter.New()
	router.ServeFiles("/assets/*filepath", http.Dir(internal.RootDir))
	router.GET("/cpanel", NewHandler(CPanelGET))

	// todo:
	// router.PanicHandler
	// router.NotFound

	router.HandlerFunc("GET", "/debug/pprof/", pprof.Index)
	router.HandlerFunc("GET", "/debug/pprof/cmdline", pprof.Cmdline)
	router.HandlerFunc("GET", "/debug/pprof/profile", pprof.Profile)
	router.HandlerFunc("GET", "/debug/pprof/symbol", pprof.Symbol)
	router.HandlerFunc("GET", "/debug/pprof/trace", pprof.Trace)

	srv := http.Server{
		Addr:    ":" + *argHTTPPort,
		Handler: router,
	}

	if internal.IsProd {
		srv.ReadHeaderTimeout = 2 * time.Second
		srv.ReadTimeout = 10 * time.Second // to allow upload photos
		srv.WriteTimeout = 10 * time.Second
	}

	// If I ever implement hijacked live lived connections
	// ie: SSE, WebSockets, ...
	// Use this callback to send a shutdown signal to those connections
	// srv.RegisterOnShutdown(f func())

	var g errgroup.Group

	g.Go(func() error {
		log.Info().Str("addr", srv.Addr).Msg("http server listening")

		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		ctxSD, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		err := srv.Shutdown(ctxSD)
		if err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}

		return nil
	})

	return g.Wait()
}
