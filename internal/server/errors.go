package server

import (
	"errors"
	"net/http"
	"os"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/rs/zerolog/log"
)

var ErrBadParam = errors.New("invalid query string param")

func errorHandler(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	lgr := log.Err(err)

	var serr internal.SErr
	if errors.As(err, &serr) {
		err = serr.Private
		lgr = lgr.Str("public", serr.Public.Error())
	}

	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrBadParam):
		code = http.StatusBadRequest
	}

	lgr.Int("status", code).Msg("http error")

	var errst internal.ErrWithStackTrace
	if errors.As(err, &errst) {
		os.Stderr.Write(errst.Stack)
	}

	http.Error(w, serr.Public.Error(), code)
}
