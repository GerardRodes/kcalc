package internal

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

func Must[T any](v T, err error) T {
	if err != nil {
		log.Panic().Err(err).Msg("must")
	}
	return v
}

func KJ2KCal(kJ float64) float64 {
	return kJ * 0.2390057361377
}

func Measure(h func()) {
	start := time.Now()
	defer func() {
		fmt.Printf("took: %v\n", time.Since(start))
	}()

	fmt.Println("start")
	h()
}
