package server

import (
	"context"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
)

type Controller func(w http.ResponseWriter, r *http.Request, p httprouter.Params) error

func NewHandler(c Controller) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// todo: handle panic
		log.Debug().Str("method", r.Method).Str("path", r.URL.Path).Msg("request")

		http.TimeoutHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Second*8)
			defer cancel()
			r = r.WithContext(ctx)
			errorHandler(w, c(w, r, p))
		}), time.Second*9, "http handler request timeout").ServeHTTP(w, r)
	}
}
