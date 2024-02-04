package internal

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/automaxprocs/maxprocs"
)

var (
	argRootDir = flag.String("root-dir", "", "")
	argLogLVL  = flag.String("log-lvl", "debug", "")
	IsProd     = false
)

func init() {
	time.Local = time.UTC
	IsProd = !strings.Contains(os.Args[0], "go-build") && !strings.HasPrefix(os.Args[0], "/tmp/")

	if IsProd {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

func Entrypoint(run func(context.Context) error) {
	flag.Parse()

	RootDir = *argRootDir

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	lvl, err := zerolog.ParseLevel(*argLogLVL)
	if err != nil {
		log.Err(err).Msg("parse zerolog lvl")
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(lvl)

	log.Print("🚀 starting")
	defer log.Print("👋 bye")

	defer func() { log.Debug().Str("root_dir", RootDir).Msg("run end") }()

	if err := entrypoint(ctx, run); err != nil {
		os.Exit(1)
	}
}

func entrypoint(
	ctx context.Context,
	run func(context.Context) error,
) (outErr error) {
	_, _ = maxprocs.Set(maxprocs.Logger(func(s string, i ...any) { // No need to handle error
		log.Printf(s, i...)
	}))

	defer func() {
		if rcv := recover(); rcv != nil {
			var err error
			if errV, ok := rcv.(error); ok {
				err = errV
			} else {
				err = fmt.Errorf("%v", rcv)
			}

			log.Err(err).Msg("recovered from panic")

			if !IsProd {
				debug.PrintStack()
			}
		} else if outErr != nil {
			log.Err(outErr).Msg("entrypoint")

			var errst ErrWithStackTrace
			if errors.As(outErr, &errst) {
				os.Stderr.Write(errst.Stack)
			}
		}
	}()

	return run(ctx)
}
