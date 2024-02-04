package server

import (
	"context"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	servertiming "github.com/mitchellh/go-server-timing"
	"github.com/rs/zerolog/log"
)

type Controller func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error

func NewHandler(c Controller) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Debug().Str("method", r.Method).Str("uri", r.URL.RequestURI()).Msg("request")

		h := http.TimeoutHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timing := servertiming.FromContext(r.Context())
			defer timing.NewMetric("handler").Start().Stop()

			ctx, cancel := context.WithTimeout(r.Context(), time.Second*8) // custom ctx timeout, lower than timeout handler
			defer cancel()
			r = r.WithContext(ctx)

			errorHandler(w, c(w, r, p))
		}), time.Second*9, "http handler request timeout")
		h = servertiming.Middleware(h, nil)

		h.ServeHTTP(w, r)
	}
}
