package ksqlite

import (
	"errors"
	"fmt"

	"github.com/GerardRodes/kcalc/internal"
)

func AddSource(src internal.Source) (int64, error) {
	sourceID, err := GetSourceID(src.Name)
	if errors.Is(err, internal.ErrNotFound) {
		return WQueryOne[int64](`
			insert into sources (name)
			values (?)
			on conflict do update
			set id = id
			returning id;
		`, src.Name)
	}

	if err != nil {
		return 0, fmt.Errorf("get source id: %w", err)
	}

	return sourceID, nil
}

func GetSourceID(name string) (int64, error) {
	return RQueryOne[int64]("select id from sources where name = ?", name)
}
