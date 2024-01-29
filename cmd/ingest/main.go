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
)

var argDBPath = flag.String("db-path", "", "")

func main() {
	internal.Entrypoint(run)
}

func run(ctx context.Context) error {
	if *argDBPath != "" {
		if err := os.MkdirAll(filepath.Dir(*argDBPath), os.ModePerm); err != nil {
			return fmt.Errorf("make dir for db: %w", err)
		}
	}

	if err := ksqlite.InitGlobals(*argDBPath, 1); err != nil {
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
