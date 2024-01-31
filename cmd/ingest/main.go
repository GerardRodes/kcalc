package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/GerardRodes/kcalc/internal/nutritionix"
	"github.com/rs/zerolog/log"
)

var argRootDir = flag.String("root-dir", "", "")

func main() {
	internal.Entrypoint(run)
}

func run(ctx context.Context) error {
	rootDir := *argRootDir
	if rootDir != "" {
		if err := os.MkdirAll(filepath.Dir(rootDir), os.ModePerm); err != nil {
			return fmt.Errorf("make dir for db: %w", err)
		}
	} else {
		var err error
		rootDir, err = os.MkdirTemp("", "kcalc_*")
		if err != nil {
			return fmt.Errorf("mkdir temp: %w", err)
		}
	}

	log.Debug().Str("root_dir", rootDir).Msg("run")
	defer log.Debug().Str("root_dir", rootDir).Msg("run end")

	if err := ksqlite.InitGlobals(filepath.Join(rootDir, "kcalc.db"), 1); err != nil {
		return fmt.Errorf("init ksqlite globals: %w", err)
	}

	err := nutritionix.Ingest(ctx, "data/ready/nutritionix.com")
	if err != nil {
		return fmt.Errorf("ingest nutritionix: %w", err)
	}

	if err := ksqlite.CloseGlobals(); err != nil {
		return fmt.Errorf("close ksqlite globals: %w", err)
	}

	return nil
}
