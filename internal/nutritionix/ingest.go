package nutritionix

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
)

type sourceFoodPhoto struct {
	Thumb   string `json:"thumb"`
	Highres string `json:"highres"`
}
type sourceFood struct {
	Name               string          `json:"food_name"`
	ServingQty         float32         `json:"serving_qty"` // qty+unit. ie:'g'+10 = 10g
	ServingUnit        string          `json:"serving_unit"`
	ServingWeightGrams float32         `json:"serving_weight_grams"`
	Calories           float32         `json:"nf_calories"`
	Photo              sourceFoodPhoto `json:"photo"`
}
type nutritionixSource struct {
	Total int
	Foods []sourceFood
}

func Ingest(ctx context.Context, jsonsDir string) error {
	children, err := os.ReadDir(jsonsDir)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	source, err := ksqlite.AddSource(internal.Source{Name: "nutritionix"})
	if err != nil {
		return fmt.Errorf("add source: %w", err)
	}

	spew.Dump(source)
	return nil

	for idx, c := range children {
		if c.IsDir() {
			continue
		}

		lgr := log.With().
			Str("name", c.Name()).
			Logger()

		if !strings.HasSuffix(mime.TypeByExtension(filepath.Ext(c.Name())), "/json") {
			lgr.Debug().
				Str("ext", filepath.Ext(c.Name())).
				Str("mime", mime.TypeByExtension(filepath.Ext(c.Name()))).
				Msg("skip")
			continue
		}

		f, err := os.Open(path.Join(jsonsDir, c.Name()))
		if err != nil {
			return fmt.Errorf("open %q: %w", c.Name(), err)
		}

		var data nutritionixSource
		if err := json.NewDecoder(f).Decode(&data); err != nil {
			return fmt.Errorf("decoding %q: %w", c.Name(), err)
		}

		// foods := make([]internal.Food, 0, len(data.Foods))
		// for _, sourceFood := range data.Foods {
		// 	food := internal.Food{
		// 		DetailsFromSources: map[int64]internal.FoodsDetail{
		// 			"nutritionix": {},
		// 		},
		// 	}

		// 	foods = append(foods, food)
		// }

		// if err := ksqlite.AddFoods(foods...); err != nil {
		// 	return fmt.Errorf("add foods: %w", err)
		// }

		log.Info().
			Int("total", len(children)).
			Int("idx", idx).
			Msg("ingested")
	}

	return nil
}
