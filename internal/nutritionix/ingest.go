package nutritionix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/fsstorage"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/gertd/go-pluralize"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type sourceFoodPhoto struct {
	// Thumb   string `json:"thumb"` is equal to highres
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

	filesGroup, ctx := errgroup.WithContext(ctx)
	filesGroup.SetLimit(runtime.NumCPU() - 1)

	pluralizeC := pluralize.NewClient()

	var done atomic.Int32
	for _, c := range children {
		c := c

		if err := ctx.Err(); err != nil {
			log.Debug().Msg("children context cancelled")
			return errors.Join(filesGroup.Wait(), fmt.Errorf("for children: %w", err))
		}

		if c.IsDir() {
			continue
		}

		lgr := log.With().
			Str("source", "nutritionix").
			Str("name", c.Name()).
			Logger()

		if !strings.HasSuffix(mime.TypeByExtension(filepath.Ext(c.Name())), "/json") {
			lgr.Debug().
				Str("ext", filepath.Ext(c.Name())).
				Str("mime", mime.TypeByExtension(filepath.Ext(c.Name()))).
				Msg("skip")
			continue
		}

		filesGroup.Go(func() error {
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
			defer f.Close()

			var data nutritionixSource
			if err := json.NewDecoder(f).Decode(&data); err != nil {
				return fmt.Errorf("decoding %q: %w", c.Name(), err)
			}

			foodGroup, ctx := errgroup.WithContext(ctx)
			foodGroup.SetLimit(1)

			for _, srcFood := range data.Foods {
				if err := ctx.Err(); err != nil {
					log.Debug().Msg("data.Foods context cancelled")
					return errors.Join(foodGroup.Wait(), fmt.Errorf("for data.Foods: %w", err))
				}

				srcFood := srcFood
				foodGroup.Go(func() error {
					var foodDetail internal.FoodDetail

					if srcFood.ServingWeightGrams > 0 {
						foodDetail.KCal = float64(srcFood.Calories) / float64(srcFood.ServingWeightGrams)
					}

					name := pluralizeC.Singular(srcFood.Name)
					food := internal.Food{
						DetailsFromSources: map[int64]internal.FoodDetail{
							sourceID: foodDetail,
						},
						ImagesFromSources: map[int64][]internal.FoodImage{},
						Locales: map[int64]internal.Locale{
							langID: {
								Value:  name,
								Normal: internal.MustNormalizeStr(name),
							},
						},
					}

					for _, v := range []string{srcFood.Photo.Highres /*srcFood.Photo.Thumb they are the same*/} {
						if v == "" {
							continue
						}

						foodImage, err := fsstorage.StoreImage(v)
						if err != nil {
							lgr.Err(err).Str("url", v).Msg("download image")
							continue
						}
						food.ImagesFromSources[sourceID] = append(food.ImagesFromSources[sourceID], foodImage)
					}

					if err := ksqlite.AddFood(food); err != nil {
						return fmt.Errorf("add food: %w", err)
					}

					return nil
				})
			}

			if err := foodGroup.Wait(); err != nil {
				return fmt.Errorf("group: %w", err)
			}

			lgr.Info().
				Int("total", len(children)).
				Int("done", int(done.Add(1))).
				Msg("ingested")

			return nil
		})
	}

	if err := filesGroup.Wait(); err != nil {
		return fmt.Errorf("files group: %w", err)
	}

	return nil
}
