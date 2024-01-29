package internal

import "context"

type RunFunc func(ctx context.Context) error

func Ingest(runners RunFunc) error {
	return nil
}
