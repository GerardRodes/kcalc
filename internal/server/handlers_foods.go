package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/fsstorage"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
)

func FoodsForm(w http.ResponseWriter, r *http.Request) error {
	return tmpl.ExecuteTemplate(w, "foods_form", newData(nil))
}

func FoodsList(w http.ResponseWriter, r *http.Request) error {
	s, err := SessionFromReq(r)
	if err != nil {
		return err
	}

	search := r.URL.Query().Get("search")
	foods, err := ksqlite.FindFoods(search)
	if err != nil {
		return internal.NewSErr(
			fmt.Errorf("bad search: %w", internal.ErrInvalid),
			fmt.Errorf("find foods: %w", err),
		)
	}

	type foodTmpl struct {
		ID    int64
		Name  string
		Image internal.Image
	}
	foodsTmpl := make([]foodTmpl, len(foods))
	for i, food := range foods {
		foodsTmpl[i].ID = food.ID
		foodsTmpl[i].Name = food.Name(s.User.Lang)

		if img, ok := food.ImageByUser[s.User.ID]; ok && img.URI != "" {
			foodsTmpl[i].Image = img
		} else {
			for _, img := range food.ImageBySource {
				foodsTmpl[i].Image = img
				break
			}
		}
	}

	return tmpl.ExecuteTemplate(w, "foods_list", newData(map[any]any{
		"foods":   foodsTmpl,
		"session": s,
		"search":  search,
	}))
}

func FoodsNew(w http.ResponseWriter, r *http.Request) error {
	const userID = 0
	if err := r.ParseMultipartForm(1024 * 4); err != nil {
		return fmt.Errorf("parse multipart form: %w", err)
	}

	food := internal.Food{
		DetailByUser: map[int64]internal.FoodDetail{},
		ImageByUser:  map[int64]internal.Image{},
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
				return internal.NewSErr(errors.New("invalid kcal format"), err)
			}
		case "g":
			g, err = strconv.ParseFloat(vals[0], 64)
			if err != nil {
				return internal.NewSErr(errors.New("invalid g format"), err)
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

			food.ImageByUser[userID] = internal.Image{URI: uri}
		}
	}

	foodID, err := ksqlite.AddFood(food)
	if err != nil {
		return fmt.Errorf("add food: %w", err)
	}

	w.Header().Add("x-up-method", "get")
	w.Header().Add("x-up-location", fmt.Sprintf("/cpanel?last_id=%d", foodID-1))
	http.Redirect(w, r, fmt.Sprintf("/cpanel?last_id=%d", foodID-1), http.StatusFound)
	return nil
}
