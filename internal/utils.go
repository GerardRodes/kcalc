package internal

import "github.com/rs/zerolog/log"

func Must[T any](v T, err error) T {
	if err != nil {
		log.Panic().Err(err).Msg("must")
	}
	return v
}
