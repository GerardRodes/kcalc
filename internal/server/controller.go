package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	servertiming "github.com/mitchellh/go-server-timing"
	"github.com/rs/zerolog/log"
)

type Controller func(w http.ResponseWriter, r *http.Request) error

func NewHandler(c Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			v := recover()
			if v == nil {
				return
			}

			err, ok := v.(error)
			if !ok {
				err = fmt.Errorf("%v", v)
			}

			log.Err(err).Msg("recovered from panic")
			w.WriteHeader(http.StatusInternalServerError)
		}()

		log.Debug().Str("method", r.Method).Str("uri", r.URL.RequestURI()).Msg("request")

		h := TimeoutMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timing := servertiming.FromContext(r.Context())
			defer timing.NewMetric("handler").Start().Stop()

			errorHandler(w, c(w, r))
		}))
		h = servertiming.Middleware(h, nil)

		h.ServeHTTP(w, r)
	}
}

func TimeoutMW(h http.Handler) http.Handler {
	return http.TimeoutHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*8) // custom ctx timeout, lower than timeout handler
		defer cancel()
		h.ServeHTTP(w, r.WithContext(ctx))
	}), time.Second*9, "http handler request timeout")
}
