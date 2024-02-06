package foodcomposition

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/rs/zerolog/log"
)

func Ingest(ctx context.Context, csvPath string) (outErr error) {
	lgr := log.With().
		Str("source", "foodcomposition").
		Logger()

	sourceID, err := ksqlite.AddSource(internal.Source{Name: "foodcomposition"})
	if err != nil {
		return fmt.Errorf("add source: %w", err)
	}

	langID, err := ksqlite.GetLang("en")
	if err != nil {
		return fmt.Errorf("get lang: %w", err)
	}

	f, err := os.OpenFile(csvPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = 37
	r.ReuseRecord = true
	r.TrimLeadingSpace = true

	_, _ = r.Read() // skip header
	var added uint64
	defer func() {
		if outErr == nil {
			lgr.Debug().Uint64("total", added).Msg("done")
		}
	}()
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read record: %w", err)
		}

		if rec[0] == "" || rec[1] == "" || rec[2] != "100" {
			continue
		}

		kJ, err := strconv.ParseFloat(rec[4], 64)
		if err != nil {
			return fmt.Errorf("parse kJ: %w", err)
		}

		_, err = ksqlite.AddFood(internal.Food{
			DetailBySource: map[int64]internal.FoodDetail{
				sourceID: {
					KCal: internal.KJ2KCal(kJ) / 100,
				},
			},
			Locales: map[int64]internal.Locale{
				langID: {
					Value: rec[1],
				},
			},
		})
		if err != nil {
			return fmt.Errorf("add food: %w", err)
		}
		added++
	}
}
