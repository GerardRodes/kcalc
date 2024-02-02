package ksqlite

import (
	"fmt"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
)

func ListFoods(lastID int64) ([]internal.Food, error) {
	ids, err := RQuery[int64](`
		select id
		from foods
		where id > ?
		order by id
		limit ?
	`, lastID, internal.PageSize)
	if err != nil {
		return nil, err
	}

	foods := make([]internal.Food, len(ids))
	for i, id := range ids {
		if err := LoadFood(id, &foods[i]); err != nil {
			return nil, err
		}
	}
	return foods, nil
}

func FindFoods(search string) ([]internal.Food, error) {
	foodIDs, err := RQuery[int64](`
		select distinct fts_fl.food_id
		from fts_foods_locales fts_fl
		where fts_fl.value match ?
		order by rank
		limit ?
	`, search, internal.PageSize)
	if err != nil {
		return nil, err
	}

	foods := make([]internal.Food, len(foodIDs))
	for i, foodID := range foodIDs {
		if err := LoadFood(foodID, &foods[i]); err != nil {
			return nil, err
		}
	}

	return foods, nil
}

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
				// todo: merge
				log.Debug().
					Int("lang_id", int(langID)).
					Str("value_normal", locale.Normal).
					Msg("food collision")
				return nil
			}
		}
	}

	var foodID int64
	if len(collisionFoodIDs) > 1 {
		// todo: merge
		log.Debug().
			Ints64("food ids", collisionFoodIDs).
			Msg("food collision")
		return nil
	} else if len(collisionFoodIDs) == 1 {
		foodID = collisionFoodIDs[0]
	} else {
		var err error
		foodID, err = WQueryOne[int64]("insert into foods default values returning id;")
		if err != nil {
			return fmt.Errorf("insert food: %w", err)
		}
	}

	for sourceID, detail := range food.DetailsFromSources {
		if err := AddFoodDetailFromSource(foodID, sourceID, detail); err != nil {
			return fmt.Errorf("food(%d) source(%d) add food detail: %w", foodID, sourceID, err)
		}
	}

	for sourceID, imgs := range food.ImagesFromSources {
		for _, img := range imgs {
			err := Exec(`
				insert into foods_images
				(food_id, source_id, height, width, uri)
				values (?, ?, ?, ?, ?);
			`, foodID, sourceID, img.Height, img.Width, img.URI)
			if err != nil {
				spew.Dump(foodID, sourceID, img.Height, img.Width, img.URI)
				return fmt.Errorf("food(%d) source(%d) add food image: %w", foodID, sourceID, err)
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

func LoadFood(foodID int64, food *internal.Food) error {
	food.ID = foodID
	if food.DetailsFromSources == nil {
		food.DetailsFromSources = map[int64]internal.FoodDetail{}
	}
	if food.DetailsFromUsers == nil {
		food.DetailsFromUsers = map[int64]internal.FoodDetail{}
	}
	if food.ImagesFromSources == nil {
		food.ImagesFromSources = map[int64][]internal.FoodImage{}
	}
	if food.ImagesFromUsers == nil {
		food.ImagesFromUsers = map[int64][]internal.FoodImage{}
	}
	if food.Locales == nil {
		food.Locales = map[int64]internal.Locale{}
	}

	{ // locales
		type rowt struct {
			LangID int64
			Value  string
		}
		rows, err := RQuery[rowt]("select lang_id, value from foods_locales where food_id = ?", foodID)
		if err != nil {
			return err
		}

		for _, row := range rows {
			food.Locales[row.LangID] = internal.Locale{Value: row.Value}
		}
	}

	{ // details
		type rowt struct {
			UserID, SourceID int64
			KCal             float64
		}
		rows, err := RQuery[rowt]("select user_id, source_id, kcal from foods_details where food_id = ?", foodID)
		if err != nil {
			return err
		}

		for _, row := range rows {
			detail := internal.FoodDetail{KCal: row.KCal}
			if row.SourceID != 0 {
				food.DetailsFromSources[row.SourceID] = detail
			} else {
				food.DetailsFromUsers[row.UserID] = detail
			}
		}
	}

	{ // images
		type rowt struct {
			UserID, SourceID int64
			URI              string
		}
		rows, err := RQuery[rowt]("select user_id, source_id, uri from foods_images where food_id = ?", foodID)
		if err != nil {
			return err
		}

		for _, row := range rows {
			img := internal.FoodImage{URI: row.URI}
			if row.SourceID != 0 {
				food.ImagesFromSources[row.SourceID] = append(food.ImagesFromSources[row.SourceID], img)
			} else {
				food.ImagesFromSources[row.UserID] = append(food.ImagesFromSources[row.UserID], img)
			}
		}
	}

	return nil
}
