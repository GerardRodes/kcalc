package server

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	_ "expvar"
	_ "net/http/pprof"
)

var (
	argHTTPPort  = flag.String("http-port", "8080", "")
	argPProfPort = flag.String("pprof-port", "8081", "")
)

//go:embed assets
var assets embed.FS

func Serve(ctx context.Context) error {
	router := http.NewServeMux()

	{ // CPANEL
		router.Handle("GET /cpanel", NewHandler(CPanelGET))
	}
	{ // FOODS
		router.Handle("GET /foods/new", NewHandler(FoodsForm))
		router.Handle("GET /foods", NewHandler(FoodsList))
		router.Handle("POST /foods", NewHandler(FoodsNew))
	}
	{ // COOKING
		router.Handle("GET /cookings", NewHandler(CookingList))
		router.Handle("POST /cookings", NewHandler(CookingNew))
		router.Handle("GET /cookings/{id}", NewHandler(CookingView))
		router.Handle("PATCH /cookings/{id}", NewHandler(CookingUpdate))
		router.Handle("GET /cookings/{id}/available-foods", NewHandler(CookingListAvailableFoods))
		router.Handle("POST /cookings/{id}/group", NewHandler(CookingGroupFoods))
		router.Handle("POST /cookings/{id}/cookings/{subCookingID}", NewHandler(CookingAddSubCooking))
	}
	{ // OTHER
		router.Handle("GET /content/",
			http.StripPrefix("/content/",
				http.FileServer(http.Dir(
					filepath.Join(internal.RootDir, "content")))))
		router.Handle("GET /assets/", http.FileServerFS(assets))
		router.Handle("/", http.RedirectHandler("/cookings", http.StatusSeeOther))
	}

	// todo:
	// not found router.Handle("/")

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

			ctxShutdown, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			err := srv.Shutdown(ctxShutdown)
			if err != nil {
				return fmt.Errorf("shutdown: %w", err)
			}

			return nil
		})
	}

	return g.Wait()
}
