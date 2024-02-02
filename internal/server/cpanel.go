package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/julienschmidt/httprouter"
)

func CPanelGET(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	var err error
	var foods []internal.Food

	q := r.URL.Query()
	if s := q.Get("search"); s != "" {
		foods, err = ksqlite.FindFoods(s)
		if err != nil {
			return fmt.Errorf("find foods: %w", err)
		}
	} else {
		var lastID int64
		if v := q.Get("last_id"); v != "" {
			lastID, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return internal.NewSErr("invalid param last_id", err, ErrBadParam)
			}
		}
		foods, err = ksqlite.ListFoods(lastID)
		if err != nil {
			return fmt.Errorf("list foods: %w", err)
		}
	}

	return tmpl.ExecuteTemplate(w, "cpanel", map[any]any{
		"foods": foods,
	})
}
