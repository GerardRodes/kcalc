package server

import (
	"fmt"
	"net/http"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
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
	return nil
}

func CookingGroupFoods(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func CookingAddCooking(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func CookingListAvailableFoods(w http.ResponseWriter, r *http.Request) error {
	s, err := SessionFromReq(r)
	if err != nil {
		return err
	}

	foods, err := ksqlite.FindCookingAvailableFoods(
		s.User.ID,
		r.PathValue("id"),
		r.URL.Query().Get("search"),
	)
	if err != nil {
		return fmt.Errorf("find foods: %w", err)
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
	}))
}
