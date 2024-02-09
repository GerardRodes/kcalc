package server

import (
	"errors"
	"net/http"
	"os"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/rs/zerolog/log"
)

var (
	ErrBadParam     = errors.New("invalid query string param")
	ErrForbidden    = errors.New(http.StatusText(http.StatusForbidden))
	ErrUnauthorized = errors.New(http.StatusText(http.StatusUnauthorized))
)

func errorHandler(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	lgr := log.Err(err)

	priv := err
	code := http.StatusInternalServerError
	pub := errors.New(http.StatusText(http.StatusInternalServerError))

	var serr internal.SErr
	if errors.As(err, &serr) {
		pub = serr.Public
		priv = serr.Private
		lgr = lgr.Err(priv).Str("public", pub.Error())
		switch {
		case errors.Is(pub, ErrBadParam):
			code = http.StatusBadRequest
		case errors.Is(pub, ErrForbidden):
			code = http.StatusForbidden
		case errors.Is(pub, ErrUnauthorized):
			code = http.StatusUnauthorized
		case errors.Is(pub, internal.ErrInvalid):
			code = http.StatusUnprocessableEntity
		}
	}

	if !internal.IsProd {
		var errst internal.ErrWithStackTrace
		if errors.As(priv, &errst) {
			os.Stderr.Write(errst.Stack)
		}
	}

	lgr.Int("status", code).Msg("http error")

	http.Error(w, pub.Error(), code)
}
