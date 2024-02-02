package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/GerardRodes/kcalc/internal/server"
	"github.com/rs/zerolog/log"
)

var argRootDir = flag.String("root-dir", "", "")

func main() {
	internal.Entrypoint(run)
}

func run(ctx context.Context) (outErr error) {
	internal.RootDir = *argRootDir
	if internal.RootDir == "" {
		return errors.New("missing root dir")
	}

	log.Debug().Str("root_dir", internal.RootDir).Msg("run")

	if err := ksqlite.InitGlobals(filepath.Join(internal.RootDir, "kcalc.db"), 1, false); err != nil {
		return fmt.Errorf("ksqlite init globals: %w", err)
	}
	defer func() {
		if err := ksqlite.CloseGlobals(); err != nil {
			outErr = errors.Join(outErr, fmt.Errorf("ksqlite close globals: %w", err))
		}
	}()

	if err := server.Serve(ctx); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}
