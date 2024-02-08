package server

import (
	"context"
	"net/http"
	"time"

	servertiming "github.com/mitchellh/go-server-timing"
	"github.com/rs/zerolog/log"
)

type Controller func(w http.ResponseWriter, r *http.Request) error

func NewHandler(c Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("method", r.Method).Str("uri", r.URL.RequestURI()).Msg("request")

		h := http.TimeoutHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timing := servertiming.FromContext(r.Context())
			defer timing.NewMetric("handler").Start().Stop()

			ctx, cancel := context.WithTimeout(r.Context(), time.Second*8) // custom ctx timeout, lower than timeout handler
			defer cancel()
			r = r.WithContext(ctx)

			errorHandler(w, c(w, r))
		}), time.Second*9, "http handler request timeout")
		h = servertiming.Middleware(h, nil)

		h.ServeHTTP(w, r)
	}
}
