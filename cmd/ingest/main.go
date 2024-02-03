package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/foodcomposition"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/GerardRodes/kcalc/internal/nutritionix"
	"github.com/rs/zerolog/log"
)

var argRootDir = flag.String("root-dir", "", "")

func main() {
	internal.Entrypoint(run)
}

func run(ctx context.Context) (outErr error) {
	rootDir := *argRootDir
	if rootDir == "" {
		var err error
		rootDir, err = os.MkdirTemp("", "kcalc_*")
		if err != nil {
			return fmt.Errorf("mkdir temp: %w", err)
		}
	}

	for _, dir := range []string{filepath.Join(rootDir, "images")} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("make dir for db: %w", err)
		}
	}

	internal.RootDir = rootDir

	log.Debug().Str("root_dir", rootDir).Msg("run")
	defer log.Debug().Str("root_dir", rootDir).Msg("run end")

	if err := ksqlite.InitGlobals(filepath.Join(rootDir, "kcalc.db"), 1, true); err != nil {
		return fmt.Errorf("init ksqlite globals: %w", err)
	}
	defer func() {
		if err := ksqlite.CloseGlobals(); err != nil {
			outErr = errors.Join(outErr, fmt.Errorf("ksqlite close globals: %w", err))
		}
	}()

	err := nutritionix.Ingest(ctx, "data/ready/nutritionix.com")
	if err != nil {
		return fmt.Errorf("ingest nutritionix: %w", err)
	}

	err = foodcomposition.Ingest(ctx, "data/ready/foodcomposition/concise-14-edition.csv")
	if err != nil {
		return fmt.Errorf("ingest foodcomposition: %w", err)
	}

	if err := ksqlite.Optimize(); err != nil {
		return fmt.Errorf("optimize: %w", err)
	}

	return nil
}
