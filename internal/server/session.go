package server

import (
	"net/http"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
)

func SessionFromReq(r *http.Request) (internal.Session, error) {
	// todo: get session from headers
	// return internal.Session{}, ErrForbbiden, ErrUnauthorized
	return internal.Session{
		User: internal.Must(ksqlite.GetUser(0)),
	}, nil
}
