package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/fsstorage"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/GerardRodes/kcalc/internal/server"
)

func main() {
	internal.Entrypoint(run)
}

func run(ctx context.Context) (outErr error) {
	if internal.RootDir == "" {
		return errors.New("missing root dir")
	}

	if err := ksqlite.InitGlobals(filepath.Join(internal.RootDir, "kcalc.db"), 1, false); err != nil {
		return fmt.Errorf("ksqlite init globals: %w", err)
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

	if err := server.Serve(ctx); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}
