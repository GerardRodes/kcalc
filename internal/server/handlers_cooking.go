package server

import (
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/davecgh/go-spew/spew"
	"github.com/segmentio/ksuid"
)

func CookingList(w http.ResponseWriter, r *http.Request) error {
	return tmpl.ExecuteTemplate(w, "cookings", nil)
}

func CookingView(w http.ResponseWriter, r *http.Request) error {
	s, err := SessionFromReq(r)
	if err != nil {
		return err
	}

	cooking, err := ksqlite.GetCooking(s.User.ID, r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("get cooking: %w", err)
	}

	return tmpl.ExecuteTemplate(w, "cooking", newData(map[any]any{
		"session": s,
		"cooking": cooking,
	}))
}

func CookingNew(w http.ResponseWriter, r *http.Request) error {
	const userID = 0

	id := ksuid.New().String()
	if err := ksqlite.NewCooking(userID, id); err != nil {
		return fmt.Errorf("new cooking: %w", err)
	}

	http.Redirect(w, r, "/cookings/"+id, http.StatusFound)
	return nil
}

func CookingUpdate(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func CookingAddFood(w http.ResponseWriter, r *http.Request) error {
	type req struct {
		ID    int64
		Name  string  `validate:"required"`
		Kcal  float64 `validate:"required"`
		G     float64 `validate:"required"`
		Photo *multipart.FileHeader
	}

	data, err := parseReq[req](r)
	if err != nil {
		return err
	}

	spew.Dump(data)
	return nil
}

func CookingGroupFoods(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func CookingAddSubCooking(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func CookingListAvailableFoods(w http.ResponseWriter, r *http.Request) error {
	s, err := SessionFromReq(r)
	if err != nil {
		return err
	}

	cookingID := r.PathValue("id")
	search := r.URL.Query().Get("search")
	foods, err := ksqlite.FindCookingAvailableFoods(s.User.ID, cookingID, search)
	if err != nil {
		return internal.NewPubErr(
			fmt.Errorf("bad search: %w", internal.ErrInvalid),
			fmt.Errorf("find foods: %w", err),
		)
	}

	type foodTmpl struct {
		ID    int64
		Name  string
		Image internal.Image
		KCal  string
	}
	foodsTmpl := make([]foodTmpl, len(foods))
	for i, food := range foods {
		foodsTmpl[i].ID = food.ID
		foodsTmpl[i].Name = food.Name(s.User.Lang)

		{ // image
			var ok bool
			foodsTmpl[i].Image, ok = food.ImageByUser[s.User.ID]

			if !ok {
				for _, img := range food.ImageBySource {
					foodsTmpl[i].Image = img
					break
				}
			}
		}

		{ // details
			details, ok := food.DetailByUser[s.User.ID]

			if !ok {
				for _, d := range food.DetailBySource {
					details = d
					break
				}
			}

			foodsTmpl[i].KCal = fmt.Sprintf("%0.2f", details.KCal)
		}
	}

	return tmpl.ExecuteTemplate(w, "cooking_available_foods", newData(map[any]any{
		"foods":     foodsTmpl,
		"session":   s,
		"search":    search,
		"cookingID": cookingID,
	}))
}
