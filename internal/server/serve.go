package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	_ "expvar"
	_ "net/http/pprof"
)

var (
	argHTTPPort  = flag.String("http-port", "8080", "")
	argPProfPort = flag.String("pprof-port", "8081", "")
)

func Serve(ctx context.Context) error {
	router := httprouter.New()
	router.ServeFiles("/assets/*filepath", http.Dir(internal.RootDir))
	router.GET("/cpanel", NewHandler(CPanelGET))
	router.GET("/foods/new", NewHandler(FoodsForm))
	router.POST("/foods", NewHandler(FoodsNew))

	// todo:
	// router.PanicHandler
	// router.NotFound

	// If I ever implement hijacked live lived connections
	// ie: SSE, WebSockets, ...
	// Use this callback to send a shutdown signal to those connections
	// srv.RegisterOnShutdown(f func())

	servers := map[string]*http.Server{
		"main": {
			Addr:              ":" + *argHTTPPort,
			ReadHeaderTimeout: 2 * time.Second,
			ReadTimeout:       10 * time.Second, // to allow upload photos
			WriteTimeout:      10 * time.Second,
			Handler:           router,
		},
	}

	if !internal.IsProd {
		servers["pprof"] = &http.Server{
			Addr:    ":" + *argPProfPort,
			Handler: http.DefaultServeMux,
		}
	}

	var g errgroup.Group
	for name, srv := range servers {
		name := name
		srv := srv

		g.Go(func() error {
			log.Info().Str("addr", srv.Addr).Str("server", name).Msg("listening")

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
	}

	return g.Wait()
}
