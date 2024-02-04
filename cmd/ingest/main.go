package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/foodcomposition"
	"github.com/GerardRodes/kcalc/internal/fsstorage"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/GerardRodes/kcalc/internal/nutritionix"
)

func main() {
	internal.Entrypoint(run)
}

func run(ctx context.Context) (outErr error) {
	if internal.RootDir == "" {
		var err error
		internal.RootDir, err = os.MkdirTemp("", "kcalc_*")
		if err != nil {
			return fmt.Errorf("mkdir temp: %w", err)
		}
	}

	for _, dir := range []string{filepath.Join(internal.RootDir, "images")} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("make dir for db: %w", err)
		}
	}

	if err := ksqlite.InitGlobals(filepath.Join(internal.RootDir, "kcalc.db"), 1, true); err != nil {
		return fmt.Errorf("init ksqlite globals: %w", err)
	}
	defer func() {
		if err := ksqlite.Optimize(); err != nil {
			outErr = errors.Join(outErr, fmt.Errorf("optimize: %w", err))
		}

		if err := ksqlite.CloseGlobals(); err != nil {
			outErr = errors.Join(outErr, fmt.Errorf("ksqlite close globals: %w", err))
		}
	}()

	if err := fsstorage.Init(); err != nil {
		return fmt.Errorf("fsstorage init: %w", err)
	}

	err := nutritionix.Ingest(ctx, "data/ready/nutritionix.com")
	if err != nil {
		return fmt.Errorf("ingest nutritionix: %w", err)
	}

	err = foodcomposition.Ingest(ctx, "data/ready/foodcomposition/concise-14-edition.csv")
	if err != nil {
		return fmt.Errorf("ingest foodcomposition: %w", err)
	}

	return nil
}
