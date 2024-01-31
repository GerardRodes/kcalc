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

	sourceID, err := ksqlite.AddSource(internal.Source{Name: "nutritionix"})
	if err != nil {
		return fmt.Errorf("add source: %w", err)
	}

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

		const prefix = len("httpswwwnutritionixcomnixapisearches")
		lang := strings.ToLower(c.Name()[prefix : prefix+2])
		switch lang {
		case "gb", "us":
			lang = "en_" + lang
		case "es", "mx":
			lang = "es_" + lang
		default:
			return fmt.Errorf("unknown lang %q", lang)
		}

		langID, err := ksqlite.GetLang(lang)
		if err != nil {
			return fmt.Errorf("get lang: %w", err)
		}

		f, err := os.Open(path.Join(jsonsDir, c.Name()))
		if err != nil {
			return fmt.Errorf("open %q: %w", c.Name(), err)
		}

		var data nutritionixSource
		if err := json.NewDecoder(f).Decode(&data); err != nil {
			return fmt.Errorf("decoding %q: %w", c.Name(), err)
		}

		foods := make([]internal.Food, 0, len(data.Foods))
		for _, srcFood := range data.Foods {
			var foodDetail internal.FoodDetail

			if srcFood.ServingWeightGrams > 0 {
				foodDetail.KCal = float64(srcFood.Calories) / float64(srcFood.ServingWeightGrams)
			}

			foods = append(foods, internal.Food{
				DetailsFromSources: map[int64][]internal.FoodDetail{
					sourceID: {foodDetail},
				},
				ImagesFromSources: map[int64][]internal.FoodImage{},
				Locales: map[int64]internal.Locale{
					langID: {
						Value:  srcFood.Name,
						Normal: internal.MustNormalizeStr(srcFood.Name),
					},
				},
			})
		}

		if err := ksqlite.AddFoods(foods...); err != nil {
			return fmt.Errorf("add foods: %w", err)
		}

		log.Info().
			Int("total", len(children)).
			Int("idx", idx).
			Msg("ingested")
	}

	return nil
}
