package ksqlite

import (
	"fmt"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/davecgh/go-spew/spew"
)

func AddFoods(foods ...internal.Food) error {
	for _, f := range foods {
		if err := AddFood(f); err != nil {
			return err
		}
	}
	return nil
}

func AddFood(food internal.Food) error {
	var collisionFoodIDs []int64
	{ // check collisions
		for langID, locale := range food.Locales {
			type foodLocaleRow struct{ FoodID, LangID int64 }
			rows, err := RQuery[foodLocaleRow](`
				select food_id, lang_id
				from foods_locales
				where value_normal like ?
					and lang_id = ?
			`, locale.Normal, langID)
			if err != nil {
				return fmt.Errorf("searching locales match: %w", err)
			}
			if len(rows) == 0 {
				continue
			} else if len(rows) == 1 {
				collisionFoodIDs = append(collisionFoodIDs, rows[0].FoodID)
			} else if len(rows) > 1 {
				// todo:
				spew.Dump(food, rows, locale)
				panic("food collision")
			}
		}
	}

	var foodID int64
	if len(collisionFoodIDs) > 1 {
		// todo:
		spew.Dump(food, collisionFoodIDs)
		panic("outer food collision")
	} else if len(collisionFoodIDs) == 1 {
		foodID = collisionFoodIDs[0]
	} else {
		var err error
		foodID, err = WQueryOne[int64]("insert into foods default values returning id;")
		if err != nil {
			return fmt.Errorf("insert food: %w", err)
		}
	}

	for sourceID, details := range food.DetailsFromSources {
		for _, detail := range details {
			if err := AddFoodDetailFromSource(foodID, sourceID, detail); err != nil {
				return fmt.Errorf("food(%d) source(%d) add food detail: %w", foodID, sourceID, err)
			}
		}
	}

	for langID, locale := range food.Locales {
		if err := AddFoodLocale(foodID, langID, locale); err != nil {
			return fmt.Errorf("add food locale: %w", err)
		}
	}

	return nil
}

func AddFoodDetailFromSource(foodID, sourceID int64, detail internal.FoodDetail) error {
	return Exec(`
		insert into foods_details
		(food_id, source_id, kcal)
		values (?, ?, ?)
		on conflict do update
		set kcal = excluded.kcal;
	`, foodID, sourceID, detail.KCal)
}

func AddFoodLocale(foodID, langID int64, locale internal.Locale) error {
	return Exec(`
		insert into foods_locales
		(food_id, lang_id, value, value_normal)
		values (?, ?, ?, ?)
		on conflict do update
		set value = excluded.value,
				value_normal = excluded.value_normal;
	`, foodID, langID, locale.Value, locale.Normal)
}
