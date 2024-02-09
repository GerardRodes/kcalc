package ksqlite

import (
	"time"

	"github.com/GerardRodes/kcalc/internal"
)

func NewCooking(userID int64, externalID string) error {
	ts := time.Now().UnixNano()
	return WExec(`
		insert into
		cookings (user_id, external_id, created_at, updated_at)
		values (?, ?, ?, ?)
	`, userID, externalID, ts, ts)
}

func GetCooking(userID int64, externalID string) (internal.Cooking, error) {
	id, err := RQueryOne[int64](`
		select id
		from cookings
		where user_id = ?
			and external_id = ?
	`, userID, externalID) // todo: accept also cookings from family
	if err != nil {
		return internal.Cooking{}, err
	}

	var cooking internal.Cooking
	return cooking, LoadCooking(id, &cooking)
}

func LoadCooking(id int64, cooking *internal.Cooking) error {
	row, err := RQueryOne[struct {
		ExternalID    string
		Name          string
		GAfterCooking float64
	}](`
		select external_id, name, g_after_cooking
		from cookings
		where id = ?
	`, id)
	if err != nil {
		return err
	}
	cooking.ID = row.ExternalID
	cooking.Name = row.Name
	cooking.GAfterCooking = row.GAfterCooking

	{ // foods
		rows, err := RQuery[struct {
			FoodID int64
			G      float64
		}](`
			select food_id, g
			from rel_cookings_foods
			where cooking_id = ?
		`, id)
		if err != nil {
			return err
		}

		rels := make([]internal.CookingFood, len(rows))
		for i := range rows {
			rels[i].G = rows[i].G
			if err := LoadFood(rows[i].FoodID, &rels[i].Food); err != nil {
				return err
			}
		}
		cooking.Foods = rels
	}

	{ // subcookings
		rows, err := RQuery[struct {
			SubCookingID int64
			G            float64
		}](`
			select sub_cooking_id, g
			from rel_cookings_cookings
			where cooking_id = ?
		`, id)
		if err != nil {
			return err
		}

		rels := make([]internal.SubCooking, len(rows))
		for i := range rows {
			rels[i].G = rows[i].G
			if err := LoadCooking(rows[i].SubCookingID, &rels[i].Cooking); err != nil {
				return err
			}
		}
		cooking.SubCookings = rels
	}

	return nil
}
