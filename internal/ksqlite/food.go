package ksqlite

import (
	"fmt"
	"time"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
)

func ListFoods(lastID int64) ([]internal.Food, int64, error) {
	ids, err := RQuery[int64](`
		select id
		from foods
		where id > ?
		order by id
		limit ?
	`, lastID, internal.PageSize)
	if err != nil {
		return nil, 0, err
	}

	foods := make([]internal.Food, len(ids))
	for i, id := range ids {
		if err := LoadFood(id, &foods[i]); err != nil {
			return nil, 0, err
		}
	}

	total, err := RQueryOne[int64]("select count(1) from foods")
	if err != nil {
		return nil, 0, err
	}
	return foods, total, nil
}

// find foods with images of user or family
// todo: func FindFoodsByUser(userID int64, search string)
func FindFoods(search string) ([]internal.Food, error) {
	if search == "" {
		return nil, nil
	}

	foodIDs, err := RQuery[int64](`
		select distinct fts_fl.food_id
		from fts_foods_locales fts_fl
		where fts_fl.value match ?
		order by rank
		limit ?
	`, search+"*", internal.PageSize*10)
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
		if _, err := AddFood(f); err != nil {
			return err
		}
	}
	return nil
}

func AddFood(food internal.Food) (int64, error) {
	var collisionFoodIDs []int64
	{ // check collisions
		for langID, locale := range food.Locales {
			normal := internal.MustNormalizeStr(locale.Value)
			type foodLocaleRow struct{ FoodID, LangID int64 }
			rows, err := RQuery[foodLocaleRow](`
				select food_id, lang_id
				from foods_locales
				where normal like ?
					and lang_id = ?
			`, normal, langID)
			if err != nil {
				return 0, fmt.Errorf("searching locales match: %w", err)
			}
			if len(rows) == 0 {
				continue
			} else if len(rows) == 1 {
				collisionFoodIDs = append(collisionFoodIDs, rows[0].FoodID)
			} else if len(rows) > 1 {
				// todo: merge
				log.Debug().
					Int("lang_id", int(langID)).
					Str("normal", normal).
					Msg("food collision")
				return 0, nil
			}
		}
	}

	var foodID int64

	err := TX(func(c *Conn) error {
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
			foodID, err = QueryOne[int64](c,
				"insert into foods (created_at) values (?) returning id;",
				time.Now().UnixNano(),
			)
			if err != nil {
				return fmt.Errorf("insert food: %w", err)
			}
		}

		{
			for sourceID, detail := range food.DetailBySource {
				err := Exec(c, `
				insert into foods_details
				(food_id, source_id, kcal)
				values (?, ?, ?)
				on conflict do update
				set kcal = excluded.kcal;
			`, foodID, sourceID, detail.KCal)
				if err != nil {
					return fmt.Errorf("food(%d) source(%d) add food detail: %w", foodID, sourceID, err)
				}
			}
			for userID, detail := range food.DetailByUser {
				err := Exec(c, `
				insert into foods_details
				(food_id, user_id, kcal)
				values (?, ?, ?)
				on conflict do update
				set kcal = excluded.kcal;
			`, foodID, userID, detail.KCal)
				if err != nil {
					return fmt.Errorf("food(%d) user(%d) add food detail: %w", foodID, userID, err)
				}
			}
		}

		{
			for sourceID, img := range food.ImageBySource {
				err := Exec(c, `
				insert into foods_images
				(food_id, source_id, kind, uri)
				values (?, ?, ?, ?);
			`, foodID, sourceID, img.Kind, img.URI)
				if err != nil {
					spew.Dump(foodID, sourceID, img)
					return fmt.Errorf("food(%d) source(%d) add food image: %w", foodID, sourceID, err)
				}
			}
			for userID, img := range food.ImageByUser {
				err := Exec(c, `
				insert into foods_images
				(food_id, user_id, kind, uri)
				values (?, ?, ?, ?);
			`, foodID, userID, img.Kind, img.URI)
				if err != nil {
					spew.Dump(foodID, userID, img)
					return fmt.Errorf("food(%d) user(%d) add food image: %w", foodID, userID, err)
				}
			}
		}

		for langID, locale := range food.Locales {
			err := Exec(c, `
			insert into foods_locales
			(food_id, lang_id, value, normal)
			values (?, ?, ?, ?)
			on conflict do update
			set value = excluded.value,
					normal = excluded.normal;
		`, foodID, langID, locale.Value, internal.MustNormalizeStr(locale.Value))
			if err != nil {
				return fmt.Errorf("add food locale: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return foodID, nil
}

func LoadFood(foodID int64, food *internal.Food) error {
	food.ID = foodID
	if food.DetailBySource == nil {
		food.DetailBySource = map[int64]internal.FoodDetail{}
	}
	if food.DetailByUser == nil {
		food.DetailByUser = map[int64]internal.FoodDetail{}
	}
	if food.ImageBySource == nil {
		food.ImageBySource = map[int64]internal.Image{}
	}
	if food.ImageByUser == nil {
		food.ImageByUser = map[int64]internal.Image{}
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
			food.Locales[row.LangID] = internal.Locale{
				Value: row.Value,
			}
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
				food.DetailBySource[row.SourceID] = detail
			} else {
				food.DetailByUser[row.UserID] = detail
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
			img := internal.Image{URI: row.URI}
			if row.SourceID != 0 {
				food.ImageBySource[row.SourceID] = img
			} else {
				food.ImageByUser[row.UserID] = img
			}
		}
	}

	return nil
}

func FindCookingAvailableFoods(userID int64, cookingID string, search string) ([]internal.Food, error) {
	if search == "" {
		return nil, nil
	}

	foodIDs, err := RQuery[int64](`
		select distinct fts_fl.food_id
		from fts_foods_locales fts_fl
		where fts_fl.value match ?
			and fts_fl.food_id not in (
						select food_id
						from rel_cookings_foods
						where cooking_id = ?)
		order by rank
		limit ?
	`, search+"*", cookingID, internal.PageSize*10)
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
