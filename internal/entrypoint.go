package internal

import (
	"context"
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
	argLogLVL = flag.String("log-lvl", "debug", "")
	IsProd    = false
)

func Entrypoint(run func(context.Context) error) {
	flag.Parse()

	time.Local = time.UTC
	IsProd = !strings.Contains(os.Args[0], "go-build")

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	if IsProd {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		lvl, err := zerolog.ParseLevel(*argLogLVL)
		if err != nil {
			log.Err(err).Msg("parse zerolog lvl")
			os.Exit(1)
		}
		zerolog.SetGlobalLevel(lvl)
	}

	log.Print("🚀 starting")
	defer log.Print("👋 bye")

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
		}
	}()

	return run(ctx)
}
