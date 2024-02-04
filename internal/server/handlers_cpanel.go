package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/julienschmidt/httprouter"
	servertiming "github.com/mitchellh/go-server-timing"
)

func CPanelGET(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	timing := servertiming.FromContext(r.Context())

	var err error
	var foods []internal.Food
	data := map[any]any{
		"langByID":   internal.LangByID,
		"sourceByID": internal.SourceByID,
	}

	queryTiming := timing.NewMetric("query").Start()
	{
		q := r.URL.Query()
		if s := q.Get("search"); s != "" {
			foods, err = ksqlite.FindFoods(s)
			if err != nil {
				return fmt.Errorf("find foods: %w", err)
			}

			data["mode"] = "search"
			data["total"] = len(foods)
		} else {
			var lastID int64
			if v := q.Get("last_id"); v != "" {
				lastID, err = strconv.ParseInt(v, 10, 64)
				if err != nil {
					return internal.NewSErr("invalid param last_id", err, ErrBadParam)
				}
			}
			var total int64
			foods, total, err = ksqlite.ListFoods(lastID)
			if err != nil {
				return fmt.Errorf("list foods: %w", err)
			}

			data["nextPageID"] = foods[len(foods)-1].ID
			data["prevPageID"] = foods[0].ID - 1 - internal.PageSize
			data["total"] = total
			data["mode"] = "list"
		}
		data["foods"] = foods
	}
	queryTiming.Stop()

	defer timing.NewMetric("template").Start().Stop()
	if r.Header.Get("X-Up-Target") == "#foods-table" {
		return tmpl.ExecuteTemplate(w, "foods_table", data)
	}

	return tmpl.ExecuteTemplate(w, "cpanel", data)
}
