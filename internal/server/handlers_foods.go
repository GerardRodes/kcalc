package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func FoodsNew(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	return tmpl.ExecuteTemplate(w, "foods_new", nil)
}
