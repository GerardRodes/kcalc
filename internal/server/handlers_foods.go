package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/fsstorage"
	"github.com/davecgh/go-spew/spew"
	"github.com/julienschmidt/httprouter"
)

func FoodsForm(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	return tmpl.ExecuteTemplate(w, "foods_form", newData())
}

func FoodsNew(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	if err := r.ParseMultipartForm(1024 * 4); err != nil {
		return fmt.Errorf("parse multipart form: %w", err)
	}

	var food internal.Food

	for name, vals := range r.MultipartForm.Value {
		switch name {
		case "locale-en":
			spew.Dump(vals)
		case "locale-es":
			spew.Dump(vals)
		case "kcal":
			spew.Dump(vals)
		case "g":
			spew.Dump(vals)
		}
	}

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

			food.ImageByUser[0 /*userID*/] = internal.FoodImage{URI: uri}
		}
	}

	return nil
}
