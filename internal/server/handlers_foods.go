package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/fsstorage"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/julienschmidt/httprouter"
)

func FoodsForm(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	return tmpl.ExecuteTemplate(w, "foods_form", newData())
}

func FoodsNew(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	const userID = 0
	if err := r.ParseMultipartForm(1024 * 4); err != nil {
		return fmt.Errorf("parse multipart form: %w", err)
	}

	food := internal.Food{
		DetailByUser: map[int64]internal.FoodDetail{},
		ImageByUser:  map[int64]internal.FoodImage{},
		Locales:      map[int64]internal.Locale{},
	}

	var err error
	var kcal, g float64
	for name, vals := range r.MultipartForm.Value {
		if len(vals) == 0 {
			continue
		}

		if vals[0] == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return nil
		}

		switch name {
		case "locale-es", "locale-en":
			food.Locales[internal.LangsID[name[len("locale-"):]]] = internal.Locale{Value: vals[0]}
		case "kcal":
			kcal, err = strconv.ParseFloat(vals[0], 64)
			if err != nil {
				return internal.NewSErr("invalid kcal format", err)
			}
		case "g":
			g, err = strconv.ParseFloat(vals[0], 64)
			if err != nil {
				return internal.NewSErr("invalid g format", err)
			}
		}
	}
	food.DetailByUser[userID] = internal.FoodDetail{KCal: kcal / g}

	if fhs, ok := r.MultipartForm.File["photo"]; ok && len(fhs) > 0 {
		for _, fh := range fhs {
			f, err := fh.Open()
			if err != nil {
				return fmt.Errorf("open file header: %w", err)
			}
			defer f.Close()

			fdata, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("read all file: %w", err)
			}

			uri, err := fsstorage.StoreImage(fdata, fh.Header.Get("content-type"))
			if err != nil {
				return fmt.Errorf("store image: %w", err)
			}

			food.ImageByUser[userID] = internal.FoodImage{URI: uri}
		}
	}

	foodID, err := ksqlite.AddFood(food)
	if err != nil {
		return fmt.Errorf("add food: %w", err)
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%d", foodID)
	return nil
}
